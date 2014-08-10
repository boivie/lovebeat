package service

import (
	"fmt"
	"regexp"
	"encoding/json"
	"github.com/hoisie/redis"
	"github.com/op/go-logging"
)

const (
	STATE_PAUSED  = "paused"
	STATE_OK      = "ok"
	STATE_WARNING = "warning"
	STATE_ERROR   = "error"
)

const (
	MAX_LOG_ENTRIES         = 1000
)

const (
	ACTION_REFRESH_VIEW = "refresh-view"
)

var (
	EMPTY_REGEXP *regexp.Regexp
)

var (
	client redis.Client
)

var log = logging.MustGetLogger("lovebeat")

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
	Name           string
	State          string
	Regexp         string
	LastUpdated    int64
	ree            *regexp.Regexp
}

type ViewCmd struct {
	Action   string
	View     string
}

func GetFromBackend(name string) *Service {
	service := &Service{
		Name: name,
		LastValue: -1,
		LastBeat: -1,
		LastUpdated: -1,
		WarningTimeout: -1,
		ErrorTimeout: -1,
		State: STATE_PAUSED,
	}

	if data, err := client.Get("lb.service." + name); err == nil {
		json.Unmarshal(data, &service)
	}
	return service
}

func (s *Service)GetExpiry(timeout int64) int64 {
	if timeout <= 0 {
		return 0
	}
	return s.LastBeat + timeout
 }

func (s *Service) StateAt(ts int64) string {
	var state = STATE_OK
	var warningExpiry = s.GetExpiry(s.WarningTimeout)
	var errorExpiry = s.GetExpiry(s.ErrorTimeout)
	if warningExpiry > 0 && ts >= warningExpiry {
		state = STATE_WARNING
	}
	if errorExpiry > 0 && ts >= errorExpiry {
		state = STATE_ERROR
	}
	return state
}

func (s *Service) Log(format string, args ...interface{}) {
	var key = "lb.service-log." + s.Name
	var log = fmt.Sprintf(format, args...)
	client.Lpush(key, []byte(log))
	client.Ltrim(key, 0, MAX_LOG_ENTRIES)
}

func (s *Service)Save(ref *Service, ts int64) {
	if *s != *ref {
		if s.State != ref.State {
			log.Info("SERVICE '%s', state %s -> %s",
				s.Name, ref.State, s.State)
			s.Log("%d|state|%s", ts, s.State)
		}
		if s.WarningTimeout != ref.WarningTimeout {
			log.Info("SERVICE '%s', warn %d -> %d",
				s.Name, ref.WarningTimeout, s.WarningTimeout)
			s.Log("%d|warn-tmo|%s", ts, ref.WarningTimeout)
		}
		if s.ErrorTimeout != ref.ErrorTimeout {
			log.Info("SERVICE '%s', err %d -> %d",
				s.Name, ref.ErrorTimeout, s.ErrorTimeout)
			s.Log("%d|err-tmo|%s", ts, ref.ErrorTimeout)
		}
		s.LastUpdated = ts
		b, _ := json.Marshal(s)
		client.Set("lb.service." + s.Name, b)
		if ref.LastUpdated < 0 {
			client.Sadd("lb.services.all", []byte(s.Name))
		}
	}
}

func (v *View) Refresh(ts int64) {
	v.State = STATE_OK
	for _, s := range services {
		if v.ree.Match([]byte(s.Name)) {
			if s.State == STATE_WARNING && v.State == STATE_OK  {
				v.State = STATE_WARNING
			} else if s.State == STATE_ERROR {
				v.State = STATE_ERROR
			}
		}
	}
}

func (v *View) Log(format string, args ...interface{}) {
	var key = "lb.view-log." + v.Name
	var log = fmt.Sprintf(format, args...)
	client.Lpush(key, []byte(log))
	client.Ltrim(key, 0, MAX_LOG_ENTRIES)
}

func (v *View) Save(ref *View, ts int64) {
	if *v != *ref {
		if v.State != ref.State {
			log.Info("VIEW '%s', state %s -> %s",
				v.Name, ref.State, v.State)
			v.Log("%d|state|%s", ts, v.State)
		}
		v.LastUpdated = ts
		b, _ := json.Marshal(v)
		client.Set("lb.view." + v.Name, b)
	}
}

func GetViewFromBackend(name string) *View {
	view := &View{
		Name: name,
		State: STATE_OK,
		Regexp: "",
		ree: EMPTY_REGEXP,
	}

	if data, err := client.Get("lb.view." + name); err == nil {
		json.Unmarshal(data, &view)
		view.ree, _ = regexp.Compile(view.Regexp)
	}
	return view
}

func (s *Service) UpdateViews(channel chan *ViewCmd) {
	for _, view := range views {
		if view.ree.Match([]byte(s.Name)) {
			channel <- &ViewCmd{
				Action: ACTION_REFRESH_VIEW,
				View:   view.Name,
			}
		}
	}
}

func GetServices() map[string]*Service {
	return services
}

func GetService(name string) *Service {
	var s, ok = services[name]
	if !ok {
		log.Error("Asked for unknown service %s", name)
		s = GetFromBackend(name)
		s.LastUpdated = -1
		services[name] = s
	}
	return s
}


func GetView(name string) *View {
	var s, ok = views[name]
	if !ok {
		log.Error("Asked for unknown view %s", name)
		s = GetViewFromBackend(name)
		s.LastUpdated = -1
		views[name] = s
	}
	return s
}


func CreateView(name string, expr string, channel chan *ViewCmd, ts int64) {
	var ree, err = regexp.Compile(expr)
	if err != nil {
		log.Error("Invalid regexp: %s", err)
		return
	}

	var view = GetView(name)
	var ref = *view
	view.Regexp = expr
	view.ree = ree
	view.Save(&ref, ts)

	log.Info("VIEW '%s' created or updated.", name)

	channel <- &ViewCmd{ Action: ACTION_REFRESH_VIEW, View: name }
}

var (
	services = make(map[string]*Service)
	views = make(map[string]*View)
)

func Startup() {
	EMPTY_REGEXP, _ = regexp.Compile("^$")
	var namesBytes, _ = client.Smembers("lb.services.all")
	for _, nameByte := range namesBytes {
		var name = string(nameByte)
		services[name] = GetFromBackend(name)
		log.Debug("Found service %s", name)
	}

	namesBytes, _ = client.Smembers("lb.views.all")
	for _, nameByte := range namesBytes {
		var name = string(nameByte)
		views[name] = GetViewFromBackend(name)
		log.Debug("Found view %s", name)
	}

}
