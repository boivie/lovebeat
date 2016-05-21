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

func (s *debugLogger) OnViewAdded(ts int64, view model.View, config config.ConfigView) {
	log.Infof("VIEW '%s', created -> %s", view.Name, view.State)
}
func (s *debugLogger) OnViewUpdated(ts int64, oldView, newView model.View, config config.ConfigView) {
	log.Infof("VIEW '%s', state %s -> %s", oldView.Name, oldView.State, newView.State)
}
func (s *debugLogger) OnViewRemoved(ts int64, view model.View, config config.ConfigView) {
	log.Infof("VIEW '%s', %s -> deleted", view.Name, view.State)
}

func NewDebugLogger() ServiceCallback {
	return &debugLogger{}
}
