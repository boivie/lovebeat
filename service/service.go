package service

import (
	"github.com/boivie/lovebeat/backend"
	"github.com/boivie/lovebeat/model"
	"time"
)

type Service struct {
	data          model.Service
	warningExpiry int64
	errorExpiry   int64
}

func newService(name string) *Service {
	return &Service{
		data: model.Service{
			Name:           name,
			LastValue:      -1,
			LastBeat:       -1,
			PreviousBeats:  make([]int64, model.PreviousBeatsCount),
			LastUpdated:    -1,
			WarningTimeout: -1,
			ErrorTimeout:   -1,
			State:          model.StatePaused,
		},
	}
}

func now() int64 { return time.Now().Unix() }

func (s *Service) updateExpiry() {
	s.warningExpiry = 0
	s.errorExpiry = 0

	if s.data.WarningTimeout > 0 {
		s.warningExpiry = s.data.LastBeat + s.data.WarningTimeout
	} else if s.data.WarningTimeout == TIMEOUT_AUTO {
		auto := calcTimeout(s.data.PreviousBeats)
		if auto != TIMEOUT_AUTO {
			s.warningExpiry = s.data.LastBeat + auto
		}
	}

	if s.data.ErrorTimeout > 0 {
		s.errorExpiry = s.data.LastBeat + s.data.ErrorTimeout
	} else if s.data.ErrorTimeout == TIMEOUT_AUTO {
		auto := calcTimeout(s.data.PreviousBeats)
		if auto != TIMEOUT_AUTO {
			s.errorExpiry = s.data.LastBeat + auto
		}
	}

	log.Debug("Warning expiry for %s = %d", s.name(), s.warningExpiry)
	log.Debug("Error expiry for %s = %d", s.name(), s.errorExpiry)
}

func (s *Service) name() string { return s.data.Name }

func (s *Service) updateState(ts int64) {
	s.data.State = s.stateAt(ts)
	s.data.LastUpdated = ts
}

func (s *Service) stateAt(ts int64) string {
	var state = model.StateOk
	if s.warningExpiry > 0 && ts >= s.warningExpiry {
		state = model.StateWarning
	}
	if s.errorExpiry > 0 && ts >= s.errorExpiry {
		state = model.StateError
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
	be.SaveService(&s.data)
}
