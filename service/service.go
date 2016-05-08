package service

import (
	"github.com/boivie/lovebeat/algorithms"
	"github.com/boivie/lovebeat/model"
)

type Service struct {
	data     model.Service
	inViews  []*View
	lastBeat int64
	expiry   int64
}

func newService(name string) *Service {
	return &Service{
		data: model.Service{
			Name:     name,
			LastBeat: -1,
			Timeout:  -1,
			State:    model.StatePaused,
		},
	}
}

func (s *Service) updateExpiry() {
	s.expiry = 0

	if s.data.Timeout > 0 {
		s.expiry = s.data.LastBeat + s.data.Timeout
	} else if s.data.Timeout == model.TIMEOUT_AUTO {
		auto := algorithms.AutoAlg(s.data.BeatHistory)
		if auto != model.TIMEOUT_AUTO {
			s.expiry = s.data.LastBeat + auto
		}
	}

	if s.expiry > 0 {
		log.Debug("Expiry for %s = %d", s.name(), s.expiry)
	}
}

func (s *Service) name() string { return s.data.Name }

func (s *Service) stateAt(ts int64) string {
	var state = model.StateOk
	if s.data.MutedSince > 0 {
		state = model.StateMuted
	} else if s.data.Timeout == 0 {
		state = model.StateError
	} else if s.expiry > 0 && ts >= s.expiry {
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
