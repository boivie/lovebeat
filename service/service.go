package service

import (
	"github.com/boivie/lovebeat-go/alert"
	"github.com/boivie/lovebeat-go/backend"
	"github.com/op/go-logging"
	"math"
	"regexp"
	"sort"
	"time"
)

const (
	MAX_UNPROCESSED_PACKETS = 1000
	EXPIRY_INTERVAL         = 1

	// Number of samples (diffs) we require to be able to
	// properly calculate an "auto" timeout
	AUTO_MIN_SAMPLES = 5
)

var (
	log          = logging.MustGetLogger("lovebeat")
	EMPTY_REGEXP = regexp.MustCompile("^$")
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

type Service struct {
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

func calcTimeout(values []int64) int64 {
	diffs := calcDiffs(values)
	if len(diffs) < AUTO_MIN_SAMPLES {
		log.Debug("AUTO-TIMEOUT: Not enough samples to calculate")
		return TIMEOUT_AUTO
	}

	ret := int64(math.Ceil(float64(median(diffs)) * 1.5))
	log.Debug("AUTO-TIMEOUT: vale calculated as %d", ret)
	return ret
}

func calcDiffs(values []int64) []int64 {
	var p []int64
	for i := 1; i < len(values); i++ {
		if values[i-1] != 0 && values[i] != 0 {
			p = append(p, values[i]-values[i-1])
		}
	}
	return p
}

type int64arr []int64

func (a int64arr) Len() int           { return len(a) }
func (a int64arr) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a int64arr) Less(i, j int) bool { return a[i] < a[j] }

func median(numbers []int64) int64 {
	sort.Sort(int64arr(numbers))
	middle := len(numbers) / 2
	result := numbers[middle]
	if len(numbers)%2 == 0 {
		result = (result + numbers[middle-1]) / 2
	}
	return result
}

// Called before saving - to update internal states
func (s *Service) update(ts int64) {
	s.data.State = s.stateAt(ts)

	if s.data.WarningTimeout == TIMEOUT_AUTO {
		s.data.WarningTimeout = calcTimeout(s.data.PreviousBeats)
	}
	if s.data.ErrorTimeout == TIMEOUT_AUTO {
		s.data.ErrorTimeout = calcTimeout(s.data.PreviousBeats)
	}
	s.data.LastUpdated = ts
}

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

func (s *Service) save(be backend.Backend, ref *Service, ts int64) {
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
	be.SaveService(&s.data)
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
		if ref.data.State == backend.STATE_OK {
			v.data.IncidentNbr += 1
		}
	}
	v.data.LastUpdated = ts
	v.svcs.be.SaveView(&v.data)
}

func (v *View) sendAlerts(ref *View, ts int64) {
	if v.data.State != ref.data.State {
		log.Info("VIEW '%s', %d: state %s -> %s",
			v.name(), v.data.IncidentNbr, ref.data.State,
			v.data.State)

		var services = make([]backend.StoredService, 0, 10)
		for _, s := range v.svcs.services {
			if (s.data.State == backend.STATE_WARNING ||
				s.data.State == backend.STATE_ERROR) &&
				v.contains(s.name()) {
				services = append(services, s.data)
				if len(services) == 10 {
					break
				}
			}
		}

		for _, a := range v.svcs.alerters {
			a.Notify(ref.data, v.data, services)
		}
	}
}

func (svcs *Services) updateViews(ts int64, serviceName string) {
	for _, view := range svcs.views {
		if view.ree.Match([]byte(serviceName)) {
			var ref = *view
			view.refresh(ts)
			view.save(&ref, ts)
			view.sendAlerts(&ref, ts)
		}
	}
}

func (svcs *Services) getService(name string) *Service {
	var s, ok = svcs.services[name]
	if !ok {
		log.Debug("Asked for unknown service %s", name)
		s = &Service{
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
		log.Debug("Asked for unknown view %s", name)
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
