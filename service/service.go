package service

import (
	"github.com/boivie/lovebeat-go/backend"
	"github.com/op/go-logging"
	"regexp"
	"time"
)

const (
	MAX_UNPROCESSED_PACKETS = 1000
	EXPIRY_INTERVAL         = 1
)

var (
	log          = logging.MustGetLogger("lovebeat")
	EMPTY_REGEXP = regexp.MustCompile("^$")
)

type Services struct {
	be              backend.Backend
	services        map[string]*Service
	views           map[string]*View
	serviceCmdChan  chan *serviceCmd
	viewCmdChan     chan *viewCmd
	getServicesChan chan *getServicesCmd
	getServiceChan  chan *getServiceCmd
	getViewsChan    chan *getViewsCmd
	getViewChan     chan *getViewCmd
}

type Service struct {
	svcs *Services
	data backend.StoredService
}

type View struct {
	svcs *Services
	data backend.StoredView
	ree  *regexp.Regexp
}

func now() int64 { return time.Now().Unix() }

func (s *Service) getExpiry(timeout int64) int64 {
	if timeout <= 0 {
		return 0
	}
	return s.data.LastBeat + timeout
}

func (s *Service) name() string { return s.data.Name }
func (v *View) name() string    { return v.data.Name }

func (s *Service) stateAt(ts int64) string {
	var state = backend.STATE_OK
	var warningExpiry = s.getExpiry(s.data.WarningTimeout)
	var errorExpiry = s.getExpiry(s.data.ErrorTimeout)
	if warningExpiry > 0 && ts >= warningExpiry {
		state = backend.STATE_WARNING
	}
	if errorExpiry > 0 && ts >= errorExpiry {
		state = backend.STATE_ERROR
	}
	return state
}

func (s *Service) registerBeat(ts int64) {
	s.data.LastBeat = ts
	s.data.PreviousBeats = append(s.data.PreviousBeats[1:], ts)
}

func (s *Service) save(ref *Service, ts int64) {
	if s.data.State != ref.data.State {
		log.Info("SERVICE '%s', state %s -> %s",
			s.name(), ref.data.State, s.data.State)
	}
	if s.data.WarningTimeout != ref.data.WarningTimeout {
		log.Info("SERVICE '%s', warn %d -> %d",
			s.name(), ref.data.WarningTimeout,
			s.data.WarningTimeout)
	}
	if s.data.ErrorTimeout != ref.data.ErrorTimeout {
		log.Info("SERVICE '%s', err %d -> %d",
			s.name(), ref.data.ErrorTimeout,
			s.data.ErrorTimeout)
	}
	s.data.LastUpdated = ts
	s.svcs.be.SaveService(&s.data)
}

func (v *View) refresh(ts int64) {
	v.data.State = backend.STATE_OK
	for _, s := range v.svcs.services {
		if v.ree.Match([]byte(s.name())) {
			if s.data.State == backend.STATE_WARNING && v.data.State == backend.STATE_OK {
				v.data.State = backend.STATE_WARNING
			} else if s.data.State == backend.STATE_ERROR {
				v.data.State = backend.STATE_ERROR
			}
		}
	}
}

func (v *View) contains(serviceName string) bool {
	return v.ree.Match([]byte(serviceName))
}

func (v *View) save(ref *View, ts int64) {
	if v.data.State != ref.data.State {
		if v.data.State != ref.data.State {
			log.Info("VIEW '%s', state %s -> %s",
				v.name(), ref.data.State, v.data.State)
		}
		v.data.LastUpdated = ts
		v.svcs.be.SaveView(&v.data)
	}
}

func (s *Service) updateViews() {
	for _, view := range s.svcs.views {
		if view.ree.Match([]byte(s.name())) {
			s.svcs.viewCmdChan <- &viewCmd{
				Action: ACTION_REFRESH_VIEW,
				View:   view.name(),
			}
		}
	}
}

func (svcs *Services) getService(name string) *Service {
	var s, ok = svcs.services[name]
	if !ok {
		log.Error("Asked for unknown service %s", name)
		s = &Service{
			svcs: svcs,
			data: backend.StoredService{
				Name:           name,
				LastValue:      -1,
				LastBeat:       -1,
				PreviousBeats:  make([]int64, backend.PREVIOUS_BEATS_COUNT),
				LastUpdated:    -1,
				WarningTimeout: -1,
				ErrorTimeout:   -1,
				State:          backend.STATE_PAUSED,
			},
		}
		svcs.services[name] = s
	}
	return s
}

func (svcs *Services) getView(name string) *View {
	var s, ok = svcs.views[name]
	if !ok {
		log.Error("Asked for unknown view %s", name)
		s = &View{
			svcs: svcs,
			data: backend.StoredView{
				Name:        name,
				State:       backend.STATE_OK,
				LastUpdated: -1,
				Regexp:      "^$",
			},
			ree: EMPTY_REGEXP}
		svcs.views[name] = s
	}
	return s
}

func (svcs *Services) createView(name string, expr string, ts int64) {
	var ree, err = regexp.Compile(expr)
	if err != nil {
		log.Error("Invalid regexp: %s", err)
		return
	}

	var view = svcs.getView(name)
	var ref = *view
	view.data.Regexp = expr
	view.ree = ree
	view.refresh(ts)
	view.save(&ref, ts)

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
				if s.data.State == backend.STATE_PAUSED || s.data.State == s.stateAt(ts) {
					continue
				}
				var ref = *s
				s.data.State = s.stateAt(ts)
				s.save(&ref, ts)
				s.updateViews()
			}
		case c := <-svcs.viewCmdChan:
			var ts = now()
			switch c.Action {
			case ACTION_REFRESH_VIEW:
				log.Debug("Refresh view %s", c.View)
				var view = svcs.getView(c.View)
				var ref = *view
				view.refresh(ts)
				view.save(&ref, ts)
			case ACTION_UPSERT_VIEW:
				log.Debug("Create or update view %s", c.View)
				svcs.createView(c.View, c.Regexp, now())
			case ACTION_DELETE_VIEW:
				log.Debug("Delete view %s", c.View)
				delete(svcs.views, c.View)
				svcs.be.DeleteView(c.View)
			}
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
		case c := <-svcs.serviceCmdChan:
			var ts = now()
			var s = svcs.getService(c.Service)
			var ref = *s
			var save = true
			switch c.Action {
			case ACTION_SET_WARN:
				s.data.WarningTimeout = int64(c.Value)
			case ACTION_SET_ERR:
				s.data.ErrorTimeout = int64(c.Value)
			case ACTION_BEAT:
				if c.Value > 0 {
					s.registerBeat(ts)
					log.Debug("Beat from %s", s.name())
				}
			case ACTION_DELETE:
				delete(s.svcs.services, s.name())
				s.svcs.be.DeleteService(s.name())
				save = false
			}
			s.data.State = s.stateAt(ts)
			if save {
				s.save(&ref, ts)
			}
			s.updateViews()
		}
	}
}

func NewServices(beiface backend.Backend) *Services {
	svcs := new(Services)
	svcs.be = beiface
	svcs.serviceCmdChan = make(chan *serviceCmd, MAX_UNPROCESSED_PACKETS)
	svcs.viewCmdChan = make(chan *viewCmd, MAX_UNPROCESSED_PACKETS)
	svcs.getServicesChan = make(chan *getServicesCmd, 5)
	svcs.getServiceChan = make(chan *getServiceCmd, 5)
	svcs.getViewsChan = make(chan *getViewsCmd, 5)
	svcs.getViewChan = make(chan *getViewCmd, 5)
	svcs.services = make(map[string]*Service)
	svcs.views = make(map[string]*View)

	for _, s := range svcs.be.LoadServices() {
		var svc = &Service{svcs: svcs, data: *s}
		if svc.data.PreviousBeats == nil || len(svc.data.PreviousBeats) != backend.PREVIOUS_BEATS_COUNT {
			svc.data.PreviousBeats = make([]int64, backend.PREVIOUS_BEATS_COUNT)
		}
		svcs.services[s.Name] = svc
	}

	for _, v := range svcs.be.LoadViews() {
		var ree, _ = regexp.Compile(v.Regexp)
		svcs.views[v.Name] = &View{svcs: svcs, data: *v, ree: ree}
	}

	return svcs
}
