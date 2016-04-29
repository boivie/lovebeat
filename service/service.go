package service

import (
	"github.com/boivie/lovebeat/algorithms"
	"github.com/boivie/lovebeat/model"
)

type Service struct {
	data          model.Service
	inViews       []*View
	lastBeat      int64
	warningExpiry int64
	errorExpiry   int64
}

func newService(name string) *Service {
	return &Service{
		data: model.Service{
			Name:           name,
			LastBeat:       -1,
			WarningTimeout: -1,
			ErrorTimeout:   -1,
			State:          model.StatePaused,
		},
	}
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
	if s.lastBeat > 0 {
		s.data.BeatHistory = append(s.data.BeatHistory, ts-s.lastBeat)
		if len(s.data.BeatHistory) > model.BeatHistoryCount {
			idx := len(s.data.BeatHistory) - model.BeatHistoryCount
			s.data.BeatHistory = s.data.BeatHistory[idx:]
		}
	}
	s.data.LastBeat = ts
	s.lastBeat = ts
}
