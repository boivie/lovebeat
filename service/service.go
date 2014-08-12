package service

import (
	"github.com/boivie/lovebeat-go/backend"
	"github.com/boivie/lovebeat-go/internal"
	"github.com/op/go-logging"
	"regexp"
	"strconv"
)

type Service struct {
	Name           string
	LastValue      int
	LastBeat       int64
	LastUpdated    int64
	WarningTimeout int64
	ErrorTimeout   int64
	State          string
}

type View struct {
	Name        string
	State       string
	Regexp      string
	LastUpdated int64
	ree         *regexp.Regexp
}

var (
	log      = logging.MustGetLogger("lovebeat")
	be       backend.Backend
	services = make(map[string]*Service)
	views    = make(map[string]*View)
)

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
	be.AddServiceLog(s.Name, ts, action, extra)
}

func (s *Service) Save(ref *Service, ts int64) {
	if *s != *ref {
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
		be.SaveService(s.stored())
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
	for _, s := range services {
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
	be.AddViewLog(v.Name, ts, action, extra)
}

func (v *View) Save(ref *View, ts int64) {
	if *v != *ref {
		if v.State != ref.State {
			log.Info("VIEW '%s', state %s -> %s",
				v.Name, ref.State, v.State)
			v.Log(ts, "state", v.State)
		}
		v.LastUpdated = ts
		be.SaveView(v.stored())
	}
}

func (s *Service) UpdateViews(channel chan *internal.ViewCmd) {
	for _, view := range views {
		if view.ree.Match([]byte(s.Name)) {
			channel <- &internal.ViewCmd{
				Action: internal.ACTION_REFRESH_VIEW,
				View:   view.Name,
			}
		}
	}
}

var (
	EMPTY_REGEXP = regexp.MustCompile("^$")
)

func GetServices() map[string]*Service {
	return services
}

func GetService(name string) *Service {
	var s, ok = services[name]
	if !ok {
		log.Error("Asked for unknown service %s", name)
		s := &Service{
			Name:           name,
			LastValue:      -1,
			LastBeat:       -1,
			LastUpdated:    -1,
			WarningTimeout: -1,
			ErrorTimeout:   -1,
			State:          backend.STATE_PAUSED,
		}
		services[name] = s
	}
	return s
}

func GetView(name string) *View {
	var s, ok = views[name]
	if !ok {
		log.Error("Asked for unknown view %s", name)
		s := &View{
			Name:        name,
			State:       backend.STATE_OK,
			LastUpdated: -1,
			Regexp:      "^$",
			ree:         EMPTY_REGEXP}
		views[name] = s
	}
	return s
}

func CreateView(name string, expr string, ts int64) {
	var ree, err = regexp.Compile(expr)
	if err != nil {
		log.Error("Invalid regexp: %s", err)
		return
	}

	var view = GetView(name)
	var ref = *view
	view.Regexp = expr
	view.ree = ree
	view.Refresh(ts)
	view.Save(&ref, ts)

	log.Info("VIEW '%s' created or updated.", name)
}

func Startup(beiface backend.Backend) {
	be = beiface

	for _, s := range be.LoadServices() {
		services[s.Name] = &Service{
			Name:           s.Name,
			LastValue:      s.LastValue,
			LastBeat:       s.LastBeat,
			LastUpdated:    s.LastUpdated,
			WarningTimeout: s.WarningTimeout,
			ErrorTimeout:   s.ErrorTimeout,
			State:          s.State}
	}

	for _, v := range be.LoadViews() {
		var ree, _ = regexp.Compile(v.Regexp)
		views[v.Name] = &View{
			Name:        v.Name,
			State:       v.State,
			Regexp:      v.Regexp,
			LastUpdated: v.LastUpdated,
			ree:         ree}
	}
}
