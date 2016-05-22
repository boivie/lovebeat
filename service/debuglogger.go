package service

import (
	"github.com/boivie/lovebeat/config"
	"github.com/boivie/lovebeat/model"
)

type debugLogger struct {
}

func (s *debugLogger) OnUpdate(ts int64, update model.Update) {}

func (s *debugLogger) OnServiceAdded(ts int64, service model.Service) {
	log.Infof("SERVICE '%s', created -> %s", service.Name, service.State)
}
func (s *debugLogger) OnServiceUpdated(ts int64, oldService, newService model.Service) {
	if oldService.State != newService.State {
		log.Infof("SERVICE '%s', state %s -> %s", oldService.Name, oldService.State, newService.State)
	}
	if oldService.Timeout != newService.Timeout {
		log.Infof("SERVICE '%s', tmo %d -> %d", oldService.Name, oldService.Timeout, newService.Timeout)
	}
}
func (s *debugLogger) OnServiceRemoved(ts int64, service model.Service) {
	log.Infof("SERVICE '%s', %s -> deleted", service.Name, service.State)
}

func (s *debugLogger) OnAlarmAdded(ts int64, alarm model.Alarm, config config.ConfigAlarm) {
	log.Infof("ALARM '%s', created -> %s", alarm.Name, alarm.State)
}
func (s *debugLogger) OnAlarmUpdated(ts int64, oldAlarm, newAlarm model.Alarm, config config.ConfigAlarm) {
	log.Infof("ALARM '%s', state %s -> %s", oldAlarm.Name, oldAlarm.State, newAlarm.State)
}
func (s *debugLogger) OnAlarmRemoved(ts int64, alarm model.Alarm, config config.ConfigAlarm) {
	log.Infof("ALARM '%s', %s -> deleted", alarm.Name, alarm.State)
}

func NewDebugLogger() ServiceCallback {
	return &debugLogger{}
}
