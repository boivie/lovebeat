package service

import (
	"github.com/boivie/lovebeat-go/backend"
	"github.com/boivie/lovebeat-go/internal"
	"github.com/op/go-logging"
	"regexp"
	"strconv"
	"time"
)

var (
	log          = logging.MustGetLogger("lovebeat")
	EMPTY_REGEXP = regexp.MustCompile("^$")
)

type Services struct {
	be       backend.Backend
	services map[string]*Service
	views    map[string]*View
}

type Service struct {
	svcs           *Services
	Name           string
	LastValue      int
	LastBeat       int64
	LastUpdated    int64
	WarningTimeout int64
	ErrorTimeout   int64
	State          string
}

type View struct {
	svcs        *Services
	Name        string
	State       string
	Regexp      string
	LastUpdated int64
	ree         *regexp.Regexp
}

func now() int64 { return time.Now().Unix() }

func (s *Service) GetExpiry(timeout int64) int64 {
	if timeout <= 0 {
		return 0
	}
	return s.LastBeat + timeout
}

func (s *Service) StateAt(ts int64) string {
	var state = backend.STATE_OK
	var warningExpiry = s.GetExpiry(s.WarningTimeout)
	var errorExpiry = s.GetExpiry(s.ErrorTimeout)
	if warningExpiry > 0 && ts >= warningExpiry {
		state = backend.STATE_WARNING
	}
	if errorExpiry > 0 && ts >= errorExpiry {
		state = backend.STATE_ERROR
	}
	return state
}

func (s *Service) Log(ts int64, action string, extra string) {
	s.svcs.be.AddServiceLog(s.Name, ts, action, extra)
}

func (s *Service) Equals(other *Service) bool {
	return s.stored() == other.stored()
}

func (v *View) Equals(other *View) bool {
	return v.stored() == other.stored()
}

func (s *Service) Save(ref *Service, ts int64) {
	if !s.Equals(ref) {
		if s.State != ref.State {
			log.Info("SERVICE '%s', state %s -> %s",
				s.Name, ref.State, s.State)
			s.Log(ts, "state", s.State)
		}
		if s.WarningTimeout != ref.WarningTimeout {
			log.Info("SERVICE '%s', warn %d -> %d",
				s.Name, ref.WarningTimeout, s.WarningTimeout)
			s.Log(ts, "warn-tmo", strconv.Itoa(int(ref.WarningTimeout)))
		}
		if s.ErrorTimeout != ref.ErrorTimeout {
			log.Info("SERVICE '%s', err %d -> %d",
				s.Name, ref.ErrorTimeout, s.ErrorTimeout)
			s.Log(ts, "err-tmo", strconv.Itoa(int(ref.ErrorTimeout)))
		}
		s.LastUpdated = ts
		s.svcs.be.SaveService(s.stored())
	}
}

func (s *Service) stored() *backend.StoredService {
	return &backend.StoredService{
		Name:           s.Name,
		LastValue:      s.LastValue,
		LastBeat:       s.LastBeat,
		LastUpdated:    s.LastUpdated,
		WarningTimeout: s.WarningTimeout,
		ErrorTimeout:   s.ErrorTimeout,
		State:          s.State,
	}
}

func (v *View) stored() *backend.StoredView {
	return &backend.StoredView{
		Name:        v.Name,
		State:       v.State,
		Regexp:      v.Regexp,
		LastUpdated: v.LastUpdated,
	}
}

func (v *View) Refresh(ts int64) {
	v.State = backend.STATE_OK
	for _, s := range v.svcs.services {
		if v.ree.Match([]byte(s.Name)) {
			if s.State == backend.STATE_WARNING && v.State == backend.STATE_OK {
				v.State = backend.STATE_WARNING
			} else if s.State == backend.STATE_ERROR {
				v.State = backend.STATE_ERROR
			}
		}
	}
}

func (v *View) Log(ts int64, action string, extra string) {
	v.svcs.be.AddViewLog(v.Name, ts, action, extra)
}

func (v *View) Save(ref *View, ts int64) {
	if !v.Equals(ref) {
		if v.State != ref.State {
			log.Info("VIEW '%s', state %s -> %s",
				v.Name, ref.State, v.State)
			v.Log(ts, "state", v.State)
		}
		v.LastUpdated = ts
		v.svcs.be.SaveView(v.stored())
	}
}

func (s *Service) UpdateViews(channel chan *internal.ViewCmd) {
	for _, view := range s.svcs.views {
		if view.ree.Match([]byte(s.Name)) {
			channel <- &internal.ViewCmd{
				Action: internal.ACTION_REFRESH_VIEW,
				View:   view.Name,
			}
		}
	}
}

func (svcs *Services) GetServices() map[string]*Service {
	return svcs.services
}

func (svcs *Services) GetService(name string) *Service {
	var s, ok = svcs.services[name]
	if !ok {
		log.Error("Asked for unknown service %s", name)
		s := &Service{
			svcs:           svcs,
			Name:           name,
			LastValue:      -1,
			LastBeat:       -1,
			LastUpdated:    -1,
			WarningTimeout: -1,
			ErrorTimeout:   -1,
			State:          backend.STATE_PAUSED,
		}
		svcs.services[name] = s
	}
	return s
}

func (svcs *Services) GetView(name string) *View {
	var s, ok = svcs.views[name]
	if !ok {
		log.Error("Asked for unknown view %s", name)
		s := &View{
			svcs:        svcs,
			Name:        name,
			State:       backend.STATE_OK,
			LastUpdated: -1,
			Regexp:      "^$",
			ree:         EMPTY_REGEXP}
		svcs.views[name] = s
	}
	return s
}

func (svcs *Services) CreateView(name string, expr string, ts int64) {
	var ree, err = regexp.Compile(expr)
	if err != nil {
		log.Error("Invalid regexp: %s", err)
		return
	}

	var view = svcs.GetView(name)
	var ref = *view
	view.Regexp = expr
	view.ree = ree
	view.Refresh(ts)
	view.Save(&ref, ts)

	log.Info("VIEW '%s' created or updated.", name)
}

func (svcs *Services) Startup(beiface backend.Backend) {
	svcs.be = beiface
	svcs.services = make(map[string]*Service)
	svcs.views = make(map[string]*View)

	for _, s := range svcs.be.LoadServices() {
		svcs.services[s.Name] = &Service{
			svcs:           svcs,
			Name:           s.Name,
			LastValue:      s.LastValue,
			LastBeat:       s.LastBeat,
			LastUpdated:    s.LastUpdated,
			WarningTimeout: s.WarningTimeout,
			ErrorTimeout:   s.ErrorTimeout,
			State:          s.State}
	}

	for _, v := range svcs.be.LoadViews() {
		var ree, _ = regexp.Compile(v.Regexp)
		svcs.views[v.Name] = &View{
			svcs:        svcs,
			Name:        v.Name,
			State:       v.State,
			Regexp:      v.Regexp,
			LastUpdated: v.LastUpdated,
			ree:         ree}
	}
}

func (svcs *Services) Monitor(serviceCmdChan chan *internal.Cmd,
	viewCmdChan chan *internal.ViewCmd, expiryInterval int64) {

	period := time.Duration(expiryInterval) * time.Second
	ticker := time.NewTicker(period)
	for {
		select {
		case <-ticker.C:
			var ts = now()
			for _, s := range svcs.GetServices() {
				if s.State == backend.STATE_PAUSED || s.State == s.StateAt(ts) {
					continue
				}
				var ref = *s
				s.State = s.StateAt(ts)
				s.Save(&ref, ts)
				s.UpdateViews(viewCmdChan)
			}
		case c := <-viewCmdChan:
			var ts = now()
			switch c.Action {
			case internal.ACTION_REFRESH_VIEW:
				log.Debug("Refresh view %s", c.View)
				var view = svcs.GetView(c.View)
				var ref = *view
				view.Refresh(ts)
				view.Save(&ref, ts)
			case internal.ACTION_UPSERT_VIEW:
				log.Debug("Create or update view %s", c.View)
				svcs.CreateView(c.View, c.Regexp, now())
			}
		case c := <-serviceCmdChan:
			var ts = now()
			var s = svcs.GetService(c.Service)
			var ref = *s
			switch c.Action {
			case internal.ACTION_SET_WARN:
				s.WarningTimeout = int64(c.Value)
			case internal.ACTION_SET_ERR:
				s.ErrorTimeout = int64(c.Value)
			case internal.ACTION_BEAT:
				if c.Value > 0 {
					s.LastBeat = ts
					var diff = ts - ref.LastBeat
					s.Log(ts, "beat", strconv.Itoa(int(diff)))
					log.Debug("Beat from %s", s.Name)
				}
			}
			s.State = s.StateAt(ts)
			s.Save(&ref, ts)
			s.UpdateViews(viewCmdChan)
		}
	}
}
