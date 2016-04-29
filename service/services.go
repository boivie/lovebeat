package service

import (
	"github.com/boivie/lovebeat/backend"
	"github.com/boivie/lovebeat/config"
	"github.com/boivie/lovebeat/eventbus"
	"github.com/boivie/lovebeat/model"
	"github.com/op/go-logging"

	"regexp"
	"time"
)

type Services struct {
	be                   backend.Backend
	bus                  *eventbus.EventBus
	services             map[string]*Service
	views                map[string]*View
	upsertServiceCmdChan chan *upsertServiceCmd
	deleteServiceCmdChan chan string
	getServicesChan      chan *getServicesCmd
	getServiceChan       chan *getServiceCmd
	getViewsChan         chan *getViewsCmd
	getViewChan          chan *getViewCmd
}

const (
	MAX_UNPROCESSED_PACKETS = 1000
	EXPIRY_INTERVAL = 1
)

var log = logging.MustGetLogger("lovebeat")

func now() int64 { return int64(time.Now().UnixNano() / 1e6) }

func (svcs *Services) updateService(ref Service, service *Service, ts int64) {
	service.data.State = service.stateAt(ts)
	svcs.be.SaveService(&service.data)
	if service.data.State != ref.data.State {
		log.Info("SERVICE '%s', state %s -> %s", service.name(), ref.data.State, service.data.State)
		svcs.bus.Publish(ServiceStateChangedEvent{service.data, ref.data.State, service.data.State})
	}
	if service.data.WarningTimeout != ref.data.WarningTimeout {
		log.Info("SERVICE '%s', warn %d -> %d", service.name(), ref.data.WarningTimeout, service.data.WarningTimeout)
	}
	if service.data.ErrorTimeout != ref.data.ErrorTimeout {
		log.Info("SERVICE '%s', err %d -> %d", service.name(), ref.data.ErrorTimeout, service.data.ErrorTimeout)
	}
}

func (svcs *Services) updateView(view *View, ts int64) {
	oldState := view.data.State
	view.data.State = view.calculateState()
	if view.data.State != oldState {
		if oldState == model.StateOk {
			view.data.IncidentNbr += 1
		}
		svcs.be.SaveView(&view.data)
		svcs.views[view.data.Name] = view

		log.Info("VIEW '%s', %d: state %s -> %s", view.name(), view.data.IncidentNbr, oldState, view.data.State)
		svcs.bus.Publish(ViewStateChangedEvent{view.data, oldState, view.data.State, view.failingServices()})
	}
}

func (svcs *Services) Monitor(cfg config.Config) {
	ticker := time.NewTicker(time.Duration(EXPIRY_INTERVAL) * time.Second)
	svcs.reload(cfg)

	for {
		select {
		case <-ticker.C:
			var ts = now()
			for _, s := range svcs.services {
				if s.data.State == model.StatePaused || s.data.State == s.stateAt(ts) {
					continue
				}
				var ref = *s
				svcs.updateService(ref, s, ts)
				for _, view := range s.inViews {
					svcs.updateView(view, ts)
				}
			}
		case c := <-svcs.getServicesChan:
			var ret []model.Service
			if view, ok := svcs.views[c.View]; ok {
				for _, s := range view.servicesInView {
					ret = append(ret, s.data)
				}
			}
			c.Reply <- ret
		case c := <-svcs.getServiceChan:
			var ret = svcs.services[c.Name]
			if ret == nil {
				c.Reply <- nil
			} else {
				c.Reply <- &ret.data
			}
		case c := <-svcs.getViewsChan:
			var ret []model.View
			for _, v := range svcs.views {
				ret = append(ret, v.data)
			}
			c.Reply <- ret
		case c := <-svcs.getViewChan:
			if ret, ok := svcs.views[c.Name]; ok {
				c.Reply <- &ret.data
			} else {
				c.Reply <- nil
			}
		case c := <-svcs.deleteServiceCmdChan:
			log.Info("SERVICE '%s', deleted", c)
			if s, ok := svcs.services[c]; ok {
				ts := now()
				delete(svcs.services, s.name())
				svcs.be.DeleteService(s.name())

				svcs.bus.Publish(ServiceRemovedEvent{s.data})

				for _, view := range s.inViews {
					view.removeService(s)
					svcs.updateView(view, ts)
				}
			}
		case c := <-svcs.upsertServiceCmdChan:
			var ts = now()
			var s, exist = svcs.services[c.Service]
			if !exist {
				log.Debug("Asked for unknown service %s", c.Service)
				s = newService(c.Service)
				svcs.services[c.Service] = s
				svcs.addServiceToMatchingViews(s)
			}
			var ref = *s

			if c.RegisterBeat {
				s.registerBeat(ts)
			}

			if c.WarningTimeout != 0 {
				s.data.WarningTimeout = c.WarningTimeout
			}
			if c.ErrorTimeout != 0 {
				s.data.ErrorTimeout = c.ErrorTimeout
			}

			if !exist {
				svcs.bus.Publish(ServiceAddedEvent{s.data})
			}

			s.updateExpiry(ts)
			svcs.updateService(ref, s, ts)
			for _, view := range s.inViews {
				svcs.updateView(view, ts)
			}
		}
	}
}

func (svcs *Services) loadViewsFromConfig(cfg config.Config) []*View {
	views := make([]*View, 0)

	views = append(views, &View{
		data: model.View{
			Name:   "all",
			Regexp: "",
			State:  model.StatePaused,
			Alerts: make([]string, 0),
		},
		ree: regexp.MustCompile(""),
	})

	for name, v := range cfg.Views {
		var ree, _ = regexp.Compile(v.Regexp)
		views = append(views, &View{
			data: model.View{
				Name:   name,
				Regexp: v.Regexp,
				State:  model.StatePaused,
				Alerts: v.Alerts,
			},
			ree: ree,
		})
	}
	return views
}

func (svcs *Services) reload(cfg config.Config) {
	ts := now()

	svcs.views = make(map[string]*View)

	views := svcs.loadViewsFromConfig(cfg)
	backendViews := svcs.be.LoadViews()
	for _, view := range views {
		if be, ok := backendViews[view.data.Name]; ok {
			view.data.State = be.State
			view.data.IncidentNbr = be.IncidentNbr
		}
		log.Info("Created view '%s' ('%s'), state = %s", view.data.Name, view.data.Regexp, view.data.State)
		svcs.views[view.data.Name] = view
	}

	svcs.services = make(map[string]*Service)

	for _, s := range svcs.be.LoadServices() {
		var svc = &Service{data: *s}
		svc.updateExpiry(ts)
		svcs.services[s.Name] = svc

		svcs.addServiceToMatchingViews(svc)
	}

	// Set initial state for views
	for _, v := range svcs.views {
		svcs.updateView(v, ts)
	}
}

func (svcs *Services) addServiceToMatchingViews(svc *Service) {
	for _, v := range svcs.views {
		if v.ree.Match([]byte(svc.data.Name)) {
			v.servicesInView = append(v.servicesInView, svc)
			svc.inViews = append(svc.inViews, v)
		}
	}
}

func NewServices(beiface backend.Backend, bus *eventbus.EventBus) *Services {
	svcs := new(Services)
	svcs.bus = bus
	svcs.be = beiface
	svcs.deleteServiceCmdChan = make(chan string, 5)
	svcs.upsertServiceCmdChan = make(chan *upsertServiceCmd, MAX_UNPROCESSED_PACKETS)
	svcs.getServicesChan = make(chan *getServicesCmd, 5)
	svcs.getServiceChan = make(chan *getServiceCmd, 5)
	svcs.getViewsChan = make(chan *getViewsCmd, 5)
	svcs.getViewChan = make(chan *getViewCmd, 5)

	return svcs
}
