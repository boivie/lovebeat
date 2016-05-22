package service

import (
	"github.com/boivie/lovebeat/config"
	"github.com/boivie/lovebeat/metrics"
	"github.com/boivie/lovebeat/model"
)

var (
	StateMap = map[string]int{
		model.StateNew:   0,
		model.StateOk:    1,
		model.StateError: 2,
		model.StateMuted: 3,
	}
)

type metricsReporter struct {
	metrics metrics.Metrics
}

func (s *metricsReporter) OnUpdate(ts int64, update model.Update) {}

func (s *metricsReporter) OnServiceAdded(ts int64, service model.Service) {}
func (s *metricsReporter) OnServiceUpdated(ts int64, oldService, newService model.Service) {
	s.metrics.SetGauge("service.state."+newService.Name, int(StateMap[newService.State]))
}
func (s *metricsReporter) OnServiceRemoved(ts int64, service model.Service) {}

func (s *metricsReporter) OnAlarmAdded(ts int64, alarm model.Alarm, config config.ConfigAlarm) {}
func (s *metricsReporter) OnAlarmUpdated(ts int64, oldAlarm, newAlarm model.Alarm, config config.ConfigAlarm) {
	s.metrics.SetGauge("alarm.state."+newAlarm.Name, int(StateMap[newAlarm.State]))
}
func (s *metricsReporter) OnAlarmRemoved(ts int64, alarm model.Alarm, config config.ConfigAlarm) {}

func NewMetricsReporter(m metrics.Metrics) ServiceCallback {
	return &metricsReporter{m}
}
