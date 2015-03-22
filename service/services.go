package service

import (
	"github.com/boivie/lovebeat/backend"
	"github.com/boivie/lovebeat/config"
	"github.com/boivie/lovebeat/eventbus"
	"github.com/boivie/lovebeat/metrics"
	"github.com/boivie/lovebeat/model"
	"github.com/op/go-logging"

	"regexp"
	"time"
)

type Services struct {
	be                   backend.Backend
	bus                  *eventbus.EventBus
	services             map[string]*Service
	views                map[string]View
	upsertServiceCmdChan chan *upsertServiceCmd
	deleteServiceCmdChan chan string
	getServicesChan      chan *getServicesCmd
	getServiceChan       chan *getServiceCmd
	getViewsChan         chan *getViewsCmd
	getViewChan          chan *getViewCmd
}

const (
	MAX_UNPROCESSED_PACKETS = 1000
	EXPIRY_INTERVAL         = 1
)

var (
	log      = logging.MustGetLogger("lovebeat")
	counters = metrics.NopMetrics()
	StateMap = map[string]int{
		model.STATE_PAUSED:  0,
		model.STATE_OK:      1,
		model.STATE_WARNING: 2,
		model.STATE_ERROR:   3,
	}
)

func (svcs *Services) updateService(ref Service, service *Service, ts int64) {
	service.update(ts)
	service.save(svcs.be, &ref, ts)
	if service.data.State != ref.data.State {
		log.Info("SERVICE '%s', state %s -> %s",
			service.name(), ref.data.State, service.data.State)
		counters.SetGauge("service.state."+service.name(), int(StateMap[service.data.State]))
		svcs.bus.Publish(ServiceStateChangedEvent{
			service.data,
			ref.data.State,
			service.data.State,
		})
	}
	if service.data.WarningTimeout != ref.data.WarningTimeout {
		log.Info("SERVICE '%s', warn %d -> %d",
			service.name(), ref.data.WarningTimeout,
			service.data.WarningTimeout)
	}
	if service.data.ErrorTimeout != ref.data.ErrorTimeout {
		log.Info("SERVICE '%s', err %d -> %d",
			service.name(), ref.data.ErrorTimeout,
			service.data.ErrorTimeout)
	}
	svcs.updateMatchingViews(ts, service.name())
}

func (svcs *Services) updateView(view View, ts int64) {
	var ref = view
	view.update(ts)
	if view.data.State != ref.data.State {
		view.save(svcs.be, &ref, ts)
		svcs.views[view.data.Name] = view

		log.Info("VIEW '%s', %d: state %s -> %s",
			view.name(), view.data.IncidentNbr, ref.data.State,
			view.data.State)
		counters.SetGauge("view.state."+view.name(), int(StateMap[view.data.State]))

		if view.hasAlert(&ref) {
			// TODO: Send alert
		}
		svcs.bus.Publish(ViewStateChangedEvent{view.data, ref.data.State, view.data.State})
	}
}

func (svcs *Services) updateMatchingViews(ts int64, serviceName string) {
	for _, view := range svcs.views {
		if view.contains(serviceName) {
			svcs.updateView(view, ts)
		}
	}
}

func (svcs *Services) getService(name string) *Service {
	var s, ok = svcs.services[name]
	if !ok {
		log.Debug("Asked for unknown service %s", name)
		s = newService(name)
		svcs.services[name] = s
	}
	return s
}

func (svcs *Services) Monitor(cfg config.Config) {
	period := time.Duration(EXPIRY_INTERVAL) * time.Second
	ticker := time.NewTicker(period)
	svcs.reload(cfg)

	for {
		select {
		case <-ticker.C:
			var ts = now()
			for _, s := range svcs.services {
				if s.data.State == model.STATE_PAUSED ||
					s.data.State == s.stateAt(ts) {
					continue
				}
				var ref = *s
				svcs.updateService(ref, s, ts)
			}
		case c := <-svcs.getServicesChan:
			var ret []model.Service
			var view, ok = svcs.views[c.View]
			if ok {
				for _, s := range svcs.services {
					if view.contains(s.name()) {
						ret = append(ret, s.data)
					}
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
			var ts = now()
			var s = svcs.getService(c)
			delete(svcs.services, s.name())
			svcs.be.DeleteService(s.name())
			svcs.updateMatchingViews(ts, s.name())
		case c := <-svcs.upsertServiceCmdChan:
			var ts = now()
			var s = svcs.getService(c.Service)
			var ref = *s

			if c.RegisterBeat {
				s.registerBeat(ts)
			}

			// Don't re-calculate 'auto' if we already have values
			if c.WarningTimeout == TIMEOUT_AUTO &&
				s.data.WarningTimeout == -1 {
				s.data.WarningTimeout = TIMEOUT_AUTO
				s.data.PreviousBeats = make([]int64, model.PREVIOUS_BEATS_COUNT)
			} else if c.WarningTimeout == TIMEOUT_CLEAR {
				s.data.WarningTimeout = TIMEOUT_CLEAR
			} else if c.WarningTimeout > 0 {
				s.data.WarningTimeout = c.WarningTimeout
			}
			if c.ErrorTimeout == TIMEOUT_AUTO &&
				s.data.ErrorTimeout == -1 {
				s.data.ErrorTimeout = TIMEOUT_AUTO
				s.data.PreviousBeats = make([]int64, model.PREVIOUS_BEATS_COUNT)
			} else if c.ErrorTimeout == TIMEOUT_CLEAR {
				s.data.ErrorTimeout = TIMEOUT_CLEAR
			} else if c.ErrorTimeout > 0 {
				s.data.ErrorTimeout = c.ErrorTimeout
			}
			svcs.updateService(ref, s, ts)
		}
	}
}

func (svcs *Services) loadViewsFromConfig(cfg config.Config) []View {
	views := make([]View, 0)

	views = append(views, View{
		services: svcs.services,
		data: model.View{
			Name:   "all",
			Regexp: "",
			State:  model.STATE_PAUSED,
		},
		ree: regexp.MustCompile(""),
	})

	for name, v := range cfg.Views {
		var ree, _ = regexp.Compile(v.Regexp)
		view := View{
			services: svcs.services,
			data: model.View{
				Name:   name,
				Regexp: v.Regexp,
				State:  model.STATE_PAUSED,
			},
			ree: ree,
		}
		views = append(views, view)
	}
	return views
}

func (svcs *Services) reload(cfg config.Config) {
	svcs.services = make(map[string]*Service)

	for _, s := range svcs.be.LoadServices() {
		var svc = &Service{data: *s}
		if svc.data.PreviousBeats == nil ||
			len(svc.data.PreviousBeats) != model.PREVIOUS_BEATS_COUNT {
			svc.data.PreviousBeats = make([]int64, model.PREVIOUS_BEATS_COUNT)
		}
		svcs.services[s.Name] = svc
	}

	views := svcs.loadViewsFromConfig(cfg)
	backendViews := svcs.be.LoadViews()

	svcs.views = make(map[string]View)
	ts := now()
	for _, view := range views {
		if be, ok := backendViews[view.data.Name]; ok {
			view.data.State = be.State
			view.data.LastUpdated = be.LastUpdated
			view.data.IncidentNbr = be.IncidentNbr
		}
		log.Info("Created view '%s' ('%s'), state = %s",
			view.data.Name, view.data.Regexp, view.data.State)
		svcs.views[view.data.Name] = view
		svcs.updateView(view, ts)
	}
}

func NewServices(beiface backend.Backend, m metrics.Metrics, bus *eventbus.EventBus) *Services {
	counters = m
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
