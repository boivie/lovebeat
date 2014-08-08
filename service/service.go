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
	LastUpdated    int64
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

func (s *Service)GetNextExpiry(ts int64) int64 {
	var next int64 = 0
	var warningExpiry = s.GetExpiry(s.WarningTimeout)
	var errorExpiry = s.GetExpiry(s.ErrorTimeout)
	if warningExpiry > 0 && warningExpiry > ts && (next == 0 || warningExpiry < next) {
		next = warningExpiry
	}
	if errorExpiry > 0 && errorExpiry > ts && (next == 0 || errorExpiry < next) {
		next = errorExpiry
	}
	return next
}

func (s *Service) UpdateState(ts int64) {
	s.State = STATE_OK
	var warningExpiry = s.GetExpiry(s.WarningTimeout)
	var errorExpiry = s.GetExpiry(s.ErrorTimeout)
	if warningExpiry > 0 && ts >= warningExpiry {
		s.State = STATE_WARNING
	}
	if errorExpiry > 0 && ts >= errorExpiry {
		s.State = STATE_ERROR
	}
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
	var services, _ = client.Smembers("lb.view-contents." + v.Name)
	v.State = STATE_OK

	for _, serv := range services {
		var service = GetService(string(serv))
		if service.State == STATE_WARNING && v.State == STATE_OK  {
			v.State = STATE_WARNING
		} else if service.State == STATE_ERROR {
			v.State = STATE_ERROR
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

func (s *Service) UpdateExpiry(ts int64) {
	if s.State != STATE_PAUSED {
		if expiry := s.GetNextExpiry(ts); expiry > 0 {
			client.Zadd("lb.expiry", []byte(s.Name), float64(expiry))
			return
		}
	}
	client.Zrem("lb.expiry", []byte(s.Name))
}


func GetView(name string) (*View, *View) {
	view := &View{
		Name: name,
		State: STATE_OK,
	}

	if data, err := client.Get("lb.view." + name); err == nil {
		json.Unmarshal(data, &view)
	}
	var ref = *view
	return view, &ref
}

func (s *Service) UpdateViews(channel chan *ViewCmd) {
	var views, _ = client.Smembers("lb.views.all")

	for _, view := range views {
		var view_name = string(view)
		var mbr, _ = client.Sismember("lb.view-contents." + view_name,
			[]byte(s.Name));
		if mbr {
			channel <- &ViewCmd{
				Action: ACTION_REFRESH_VIEW,
				View:   view_name,
			}
		}
	}
}


func GetServiceNames() []string {
	var namesBytes, _ = client.Smembers("lb.services.all")

	var names = make([]string, len(namesBytes))
	for idx, elem := range namesBytes {
		names[idx] = string(elem)
	}
	return names
}


func GetExpired(ts int64) []*Service {
	names, err := client.Zrangebyscore("lb.expiry", 0, float64(ts))
	if err != nil {
		log.Error("Failed to get expired services: %s", err)
		return make([]*Service, 0)
	}

	var services = make([]*Service, len(names))
	for idx, elem := range names {
		var s = GetService(string(elem))
		services[idx] = s
	}
	return services
}

func GetServices() map[string]*Service {
	return services
}

func GetService(name string) *Service {
	var s, ok = services[name]
	if !ok {
		log.Error("Asked for unknown service %s", name)
		s = GetFromBackend(name)
	}
	return s
}

func CreateView(view_name string, re *regexp.Regexp, channel chan *ViewCmd) {
	var key = "lb.view-contents." + view_name
	client.Del(key)
	var names, _ = client.Smembers("lb.services.all")
	for _, name := range names {
		if re.Match(name) {
			client.Sadd(key, name)
		}
	}

	client.Sadd("lb.views.all", []byte(view_name))
	log.Info("VIEW '%s' created or updated.", view_name)
	channel <- &ViewCmd{
		Action: ACTION_REFRESH_VIEW,
		View:   view_name,
	}

}

var (
	services = make(map[string]*Service)
	views = make(map[string]*View)
)

func Startup() {
	for _, name := range GetServiceNames() {
		services[name] = GetFromBackend(name)
		log.Debug("Found service %s", name)
	}
}
