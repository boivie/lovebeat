package service

import (
	"github.com/boivie/lovebeat/model"
	"regexp"
)


type View struct {
	servicesInView []*Service
	data           model.View
	ree            *regexp.Regexp
}

func (v *View) name() string { return v.data.Name }

func (v *View) calculateState() string {
	state := model.StateOk
	for _, s := range v.servicesInView {
		if s.data.State == model.StateWarning && state == model.StateOk {
			state = model.StateWarning
		} else if s.data.State == model.StateError {
			state= model.StateError
		}
	}
	return state
}

func (v *View) failingServices() []model.Service {
	var failedServices = make([]model.Service, 0)
	for _, s := range v.servicesInView {
		if (s.data.State == model.StateError || s.data.State == model.StateWarning) {
			failedServices = append(failedServices, s.data)
		}
	}
	return failedServices
}

func (v *View) removeService(service *Service) {
	services := v.servicesInView[:0]
	for _, x := range v.servicesInView {
		if x != service {
			services = append(services, x)
		}
	}
	v.servicesInView = services
}