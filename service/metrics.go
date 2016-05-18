package service

import (
	"github.com/boivie/lovebeat/eventbus"
	"github.com/boivie/lovebeat/metrics"
	"github.com/boivie/lovebeat/model"
)

var (
	counters metrics.Metrics
	StateMap = map[string]int{
		model.StatePaused: 0,
		model.StateOk:     1,
		model.StateError:  2,
		model.StateMuted:  3,
	}
)

func ServiceStateChanged(ev model.ServiceStateChangedEvent) {
	service := ev.Service
	counters.SetGauge("service.state."+service.Name, int(StateMap[service.State]))
}

func ViewStateChanged(ev model.ViewStateChangedEvent) {
	view := ev.View
	counters.SetGauge("view.state."+view.Name, int(StateMap[view.State]))
}

func RegisterMetrics(bus *eventbus.EventBus, m metrics.Metrics) {
	counters = m
	bus.RegisterHandler(ServiceStateChanged)
	bus.RegisterHandler(ViewStateChanged)
}
