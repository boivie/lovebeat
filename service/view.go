package service

import (
	"github.com/boivie/lovebeat/model"
	"regexp"
	"strings"
	"github.com/boivie/lovebeat/config"
)


// Instantiated from a view template
type View struct {
	servicesInView []*Service
	data           model.View
	tmpl           *ViewTemplate
}

type ViewTemplate struct {
	config config.ConfigView
	ree    *regexp.Regexp
}

var VAR_RE = regexp.MustCompile("\\\\\\$([a-z]+)")

func makePattern(p string) string {
	pattern := "^" + regexp.QuoteMeta(p) + "$"
	pattern = strings.Replace(pattern, "\\*", ".*", -1)
	pattern = VAR_RE.ReplaceAllString(pattern, "(?P<$1>[^\\.]+)")
	return pattern
}

var NAME_RE = regexp.MustCompile("\\$([a-z]+)")

func expandName(p *regexp.Regexp, serviceName, namePattern string) string {
	groups := make(map[string]string)
	values := p.FindStringSubmatch(serviceName)
	for i, field := range p.SubexpNames() {
		if i != 0 {
			groups[field] = values[i]
		}
	}

	return NAME_RE.ReplaceAllStringFunc(namePattern, func(field string) string {
		if val, ok := groups[field[1:]]; ok {
			return val
		}
		return field[1:]
	})
}

func (v *View) name() string { return v.data.Name }

func (v *View) calculateState() string {
	state := model.StateOk
	for _, s := range v.servicesInView {
		if s.data.State == model.StateError {
			state = model.StateError
		}
	}
	return state
}

func (v *View) failingServices() []model.Service {
	var failedServices = make([]model.Service, 0)
	for _, s := range v.servicesInView {
		if s.data.State == model.StateError {
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