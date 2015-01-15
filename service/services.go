package service

import (
	"github.com/boivie/lovebeat-go/alert"
	"github.com/boivie/lovebeat-go/backend"
	"regexp"
	"time"
)

type Services struct {
	be                   backend.Backend
	alerters             []alert.Alerter
	services             map[string]*Service
	views                map[string]*View
	beatCmdChan          chan string
	upsertServiceCmdChan chan *upsertServiceCmd
	deleteServiceCmdChan chan string
	deleteViewCmdChan    chan string
	upsertViewCmdChan    chan *upsertViewCmd
	getServicesChan      chan *getServicesCmd
	getServiceChan       chan *getServiceCmd
	getViewsChan         chan *getViewsCmd
	getViewChan          chan *getViewCmd
}

const (
	MAX_UNPROCESSED_PACKETS = 1000
	EXPIRY_INTERVAL         = 1
)

func (svcs *Services) sendAlert(a alert.Alert) {
	for _, alerter := range svcs.alerters {
		alerter.Notify(a)
	}

}

func (svcs *Services) updateViews(ts int64, serviceName string) {
	for _, view := range svcs.views {
		if view.contains(serviceName) {
			var ref = *view
			view.update(ts)
			if view.data.State != ref.data.State {
				view.save(svcs.be, &ref, ts)
				if view.hasAlert(&ref) {
					svcs.sendAlert(view.getAlert(&ref))
				}
			}
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

func (svcs *Services) getView(name string) *View {
	var s, ok = svcs.views[name]
	if !ok {
		log.Debug("Asked for unknown view %s", name)
		s = newView(svcs.services, name)
		svcs.views[name] = s
	}
	return s
}

func (svcs *Services) createView(name string, expr string, alertMail string,
	webhooks string, ts int64) {
	var ree, err = regexp.Compile(expr)
	if err != nil {
		log.Error("Invalid regexp: %s", err)
		return
	}

	var view = svcs.getView(name)
	var ref = *view
	view.data.Regexp = expr
	view.ree = ree
	view.data.AlertMail = alertMail
	view.data.Webhooks = webhooks
	view.update(ts)
	view.save(svcs.be, &ref, ts)

	log.Info("VIEW '%s' created or updated.", name)
}

func (svcs *Services) Monitor() {
	period := time.Duration(EXPIRY_INTERVAL) * time.Second
	ticker := time.NewTicker(period)
	for {
		select {
		case <-ticker.C:
			var ts = now()
			for _, s := range svcs.services {
				if s.data.State == backend.STATE_PAUSED ||
					s.data.State == s.stateAt(ts) {
					continue
				}
				var ref = *s
				s.update(ts)
				s.save(svcs.be, &ref, ts)
				svcs.updateViews(ts, s.name())
			}
		case c := <-svcs.upsertViewCmdChan:
			log.Debug("Create or update view %s", c.View)
			svcs.createView(c.View, c.Regexp, c.AlertMail,
				c.Webhooks, now())
		case c := <-svcs.deleteViewCmdChan:
			log.Debug("Delete view %s", c)
			delete(svcs.views, c)
			svcs.be.DeleteView(c)
		case c := <-svcs.getServicesChan:
			var ret []backend.StoredService
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
			c.Reply <- ret.data
		case c := <-svcs.getViewsChan:
			var ret []backend.StoredView
			for _, v := range svcs.views {
				ret = append(ret, v.data)
			}
			c.Reply <- ret
		case c := <-svcs.getViewChan:
			var ret = svcs.views[c.Name]
			c.Reply <- ret.data
		case c := <-svcs.beatCmdChan:
			var ts = now()
			var s = svcs.getService(c)
			var ref = *s
			s.registerBeat(ts)
			log.Debug("Beat from %s", s.name())
			s.update(ts)
			s.save(svcs.be, &ref, ts)
			svcs.updateViews(ts, s.name())
		case c := <-svcs.deleteServiceCmdChan:
			var ts = now()
			var s = svcs.getService(c)
			delete(svcs.services, s.name())
			svcs.be.DeleteService(s.name())
			svcs.updateViews(ts, s.name())
		case c := <-svcs.upsertServiceCmdChan:
			var ts = now()
			var s = svcs.getService(c.Service)
			var ref = *s
			// Don't re-calculate 'auto' if we already have values
			if c.WarningTimeout == TIMEOUT_AUTO &&
				s.data.WarningTimeout == -1 {
				s.data.WarningTimeout = TIMEOUT_AUTO
				s.data.PreviousBeats = make([]int64, backend.PREVIOUS_BEATS_COUNT)
			} else if c.WarningTimeout == TIMEOUT_CLEAR {
				s.data.WarningTimeout = TIMEOUT_CLEAR
			} else if c.WarningTimeout > 0 {
				s.data.WarningTimeout = c.WarningTimeout
			}
			if c.ErrorTimeout == TIMEOUT_AUTO &&
				s.data.ErrorTimeout == -1 {
				s.data.ErrorTimeout = TIMEOUT_AUTO
				s.data.PreviousBeats = make([]int64, backend.PREVIOUS_BEATS_COUNT)
			} else if c.ErrorTimeout == TIMEOUT_CLEAR {
				s.data.ErrorTimeout = TIMEOUT_CLEAR
			} else if c.ErrorTimeout > 0 {
				s.data.ErrorTimeout = c.ErrorTimeout
			}
			s.update(ts)
			s.save(svcs.be, &ref, ts)
			svcs.updateViews(ts, s.name())
		}
	}
}

func NewServices(beiface backend.Backend, alerters []alert.Alerter) *Services {
	svcs := new(Services)
	svcs.be = beiface
	svcs.alerters = alerters
	svcs.beatCmdChan = make(chan string, MAX_UNPROCESSED_PACKETS)
	svcs.deleteServiceCmdChan = make(chan string, 5)
	svcs.upsertServiceCmdChan = make(chan *upsertServiceCmd, 5)
	svcs.deleteViewCmdChan = make(chan string, 5)
	svcs.upsertViewCmdChan = make(chan *upsertViewCmd, 5)
	svcs.getServicesChan = make(chan *getServicesCmd, 5)
	svcs.getServiceChan = make(chan *getServiceCmd, 5)
	svcs.getViewsChan = make(chan *getViewsCmd, 5)
	svcs.getViewChan = make(chan *getViewCmd, 5)
	svcs.services = make(map[string]*Service)
	svcs.views = make(map[string]*View)

	for _, s := range svcs.be.LoadServices() {
		var svc = &Service{data: *s}
		if svc.data.PreviousBeats == nil ||
			len(svc.data.PreviousBeats) != backend.PREVIOUS_BEATS_COUNT {
			svc.data.PreviousBeats = make([]int64, backend.PREVIOUS_BEATS_COUNT)
		}
		svcs.services[s.Name] = svc
	}

	for _, v := range svcs.be.LoadViews() {
		var ree, _ = regexp.Compile(v.Regexp)
		svcs.views[v.Name] = &View{services: svcs.services, data: *v, ree: ree}
	}

	return svcs
}
