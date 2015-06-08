package service

import (
	"github.com/boivie/lovebeat/algorithms"
	"github.com/boivie/lovebeat/backend"
	"github.com/boivie/lovebeat/model"
	"time"
)

type Service struct {
	data            model.Service
	sessionLastBeat int64
	warningExpiry   int64
	errorExpiry     int64
}

func newService(name string) *Service {
	return &Service{
		data: model.Service{
			Name:           name,
			LastValue:      -1,
			LastBeat:       -1,
			WarningTimeout: -1,
			ErrorTimeout:   -1,
			State:          model.StatePaused,
		},
	}
}

func now() int64 { return int64(time.Now().UnixNano() / 1e6) }

func calcTimeout(values []int64) int64 {
	return TIMEOUT_AUTO
}

func (s *Service) updateExpiry(ts int64) {
	s.warningExpiry = 0
	s.errorExpiry = 0

	if s.data.WarningTimeout > 0 {
		s.warningExpiry = s.data.LastBeat + s.data.WarningTimeout
	} else if s.data.WarningTimeout == TIMEOUT_AUTO {
		auto := algorithms.AutoAlg(s.data.BeatHistory)
		if auto != TIMEOUT_AUTO {
			s.warningExpiry = s.data.LastBeat + auto
		}
	}

	if s.data.ErrorTimeout > 0 {
		s.errorExpiry = s.data.LastBeat + s.data.ErrorTimeout
	} else if s.data.ErrorTimeout == TIMEOUT_AUTO {
		auto := algorithms.AutoAlg(s.data.BeatHistory)
		if auto != TIMEOUT_AUTO {
			s.errorExpiry = s.data.LastBeat + auto
		}
	}

	if s.warningExpiry > 0 {
		log.Debug("Warning expiry for %s = %d (%d ms)", s.name(), s.warningExpiry, s.warningExpiry-ts)
	}
	if s.errorExpiry > 0 {
		log.Debug("Error expiry for %s = %d (%d ms)", s.name(), s.errorExpiry, s.errorExpiry-ts)
	}
}

func (s *Service) name() string { return s.data.Name }

func (s *Service) updateState(ts int64) {
	s.data.State = s.stateAt(ts)
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
		log.Debug("Beat from %s (prev %d ms ago)", s.name(), ts-s.data.LastBeat)
	} else {
		log.Debug("Beat from %s (first!)", s.name())
	}
	if s.sessionLastBeat > 0 {
		s.data.BeatHistory = append(s.data.BeatHistory, ts-s.sessionLastBeat)
		if len(s.data.BeatHistory) > model.BeatHistoryCount {
			idx := len(s.data.BeatHistory) - model.BeatHistoryCount
			s.data.BeatHistory = s.data.BeatHistory[idx:]
		}
	}
	s.data.LastBeat = ts
	s.sessionLastBeat = ts
}

func (s *Service) save(be backend.Backend, ref *Service, ts int64) {
	be.SaveService(&s.data)
}
