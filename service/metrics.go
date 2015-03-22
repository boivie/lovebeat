package service

import (
	"github.com/boivie/lovebeat/eventbus"
	"github.com/boivie/lovebeat/metrics"
	"github.com/boivie/lovebeat/model"
)

var (
	counters metrics.Metrics
	StateMap = map[string]int{
		model.STATE_PAUSED:  0,
		model.STATE_OK:      1,
		model.STATE_WARNING: 2,
		model.STATE_ERROR:   3,
	}
)

func ServiceStateChanged(ev ServiceStateChangedEvent) {
	service := ev.Service
	counters.SetGauge("service.state."+service.Name, int(StateMap[service.State]))
}

func ViewStateChanged(ev ViewStateChangedEvent) {
	view := ev.View
	counters.SetGauge("view.state."+view.Name, int(StateMap[view.State]))
}

func RegisterMetrics(bus *eventbus.EventBus, m metrics.Metrics) {
	counters = m
	bus.RegisterHandler(ServiceStateChanged)
	bus.RegisterHandler(ViewStateChanged)
}
