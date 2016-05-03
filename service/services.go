package service

import (
	"github.com/boivie/lovebeat/backend"
	"github.com/boivie/lovebeat/config"
	"github.com/boivie/lovebeat/eventbus"
	"github.com/boivie/lovebeat/model"
	"github.com/op/go-logging"

	"github.com/boivie/lovebeat/alert"
	"github.com/boivie/lovebeat/notify"
	"regexp"
	"time"
)

type Services struct {
	be       backend.Backend
	bus      *eventbus.EventBus
	alerter  alert.Alerter
	services map[string]*Service

	views         map[string]*View
	viewTemplates []ViewTemplate
	viewStates    []*model.View

	upsertServiceCmdChan chan *upsertServiceCmd
	deleteServiceCmdChan chan string
	getServicesChan      chan *getServicesCmd
	getServiceChan       chan *getServiceCmd
	getViewsChan         chan *getViewsCmd
	getViewChan          chan *getViewCmd
}

const (
	MAX_UNPROCESSED_PACKETS = 1000
)

var log = logging.MustGetLogger("lovebeat")

func now() int64 { return int64(time.Now().UnixNano() / 1e6) }

func (svcs *Services) updateService(ref Service, service *Service, ts int64) {
	service.data.State = service.stateAt(ts)
	svcs.be.SaveService(&service.data)
	if service.data.State != ref.data.State {
		log.Info("SERVICE '%s', state %s -> %s", service.name(), ref.data.State, service.data.State)
		svcs.bus.Publish(model.ServiceStateChangedEvent{service.data, ref.data.State, service.data.State})
	}
	if service.data.Timeout != ref.data.Timeout {
		log.Info("SERVICE '%s', tmo %d -> %d", service.name(), ref.data.Timeout, service.data.Timeout)
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
		svcs.bus.Publish(model.ViewStateChangedEvent{view.data, oldState, view.data.State, view.failingServices()})
		svcs.alerter.Notify(alert.AlertInfo{view.data, oldState, view.data.State, view.failingServices(), view.tmpl.config})
	}
}

func (svcs *Services) Monitor(cfg config.Config, notifier notify.Notifier) {
	updateServicesTimer := time.NewTicker(time.Duration(1) * time.Second)
	notifyTimer := time.NewTicker(time.Duration(60) * time.Second)
	svcs.reload(cfg)

	for {
		select {
		case <-updateServicesTimer.C:
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
		case <-notifyTimer.C:
			notifier.Notify("monitor")
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

				svcs.bus.Publish(model.ServiceRemovedEvent{s.data})

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
				svcs.addViewsToService(s)
			}
			var ref = *s

			if c.RegisterBeat {
				s.registerBeat(ts)
			}

			if c.HasTimeout {
				s.data.Timeout = c.Timeout
			}

			if !exist {
				svcs.bus.Publish(model.ServiceAddedEvent{s.data})
			}

			s.updateExpiry(ts)
			svcs.updateService(ref, s, ts)
			for _, view := range s.inViews {
				svcs.updateView(view, ts)
			}
		}
	}
}

func (svcs *Services) loadViewTemplates(cfg config.Config) []ViewTemplate {
	views := make([]ViewTemplate, 0)

	views = append(views, ViewTemplate{config.ConfigView{Name: "all"}, regexp.MustCompile("")})

	for _, v := range cfg.Views {
		ree, _ := regexp.Compile(makePattern(v.Pattern))
		views = append(views, ViewTemplate{v, ree})
	}
	return views
}

func (svcs *Services) reload(cfg config.Config) {
	ts := now()

	svcs.views = make(map[string]*View)

	svcs.viewTemplates = svcs.loadViewTemplates(cfg)
	svcs.viewStates = svcs.be.LoadViews()

	svcs.services = make(map[string]*Service)

	for _, s := range svcs.be.LoadServices() {
		var svc = &Service{data: *s}
		svc.updateExpiry(ts)
		svcs.services[s.Name] = svc

		svcs.addViewsToService(svc)
	}

	// Set initial state for views
	for _, v := range svcs.views {
		svcs.updateView(v, ts)
	}
}

func (svcs *Services) addViewsToService(svc *Service) {
	for _, tmpl := range svcs.viewTemplates {
		if tmpl.ree.Match([]byte(svc.data.Name)) {
			name := expandName(tmpl.ree, svc.data.Name, tmpl.config.Name)
			v, exists := svcs.views[name]
			if !exists {
				v = &View{tmpl: &tmpl, data: model.View{name, model.StatePaused, 0}}
				for _, sv := range svcs.viewStates {
					if sv.Name == name {
						v.data = *sv
					}
				}
				svcs.views[name] = v
				log.Info("Created view %s from %s", v.data.Name, tmpl.config.Name)
			}
			v.servicesInView = append(v.servicesInView, svc)
			svc.inViews = append(svc.inViews, v)
		}
	}
}

func NewServices(beiface backend.Backend, bus *eventbus.EventBus, alerter alert.Alerter) *Services {
	svcs := new(Services)
	svcs.bus = bus
	svcs.alerter = alerter
	svcs.be = beiface
	svcs.deleteServiceCmdChan = make(chan string, 5)
	svcs.upsertServiceCmdChan = make(chan *upsertServiceCmd, MAX_UNPROCESSED_PACKETS)
	svcs.getServicesChan = make(chan *getServicesCmd, 5)
	svcs.getServiceChan = make(chan *getServiceCmd, 5)
	svcs.getViewsChan = make(chan *getViewsCmd, 5)
	svcs.getViewChan = make(chan *getViewCmd, 5)

	return svcs
}
