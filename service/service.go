package service

import (
	"github.com/boivie/lovebeat-go/backend"
	"github.com/op/go-logging"
	"time"
)

var (
	log = logging.MustGetLogger("lovebeat")
)

type Service struct {
	data backend.StoredService
}

func newService(name string) *Service {
	return &Service{
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
}

func now() int64 { return time.Now().Unix() }

func (s *Service) getExpiry(timeout int64) int64 {
	if timeout <= 0 {
		return 0
	}
	return s.data.LastBeat + timeout
}

func (s *Service) name() string { return s.data.Name }

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
	if s.data.LastBeat > 0 {
		log.Debug("Beat from %s (prev %d secs ago)", s.name(), ts-s.data.LastBeat)
	} else {
		log.Debug("Beat from %s (first!)", s.name())
	}
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
