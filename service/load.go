package service

import (
	"github.com/boivie/lovebeat/backend"
	"github.com/boivie/lovebeat/config"
	"regexp"
)

func loadViewTemplates(cfg config.Config) []ViewTemplate {
	views := []ViewTemplate{
		ViewTemplate{
			config:   config.ConfigView{Name: "all"},
			includes: []*regexp.Regexp{regexp.MustCompile("")}},
	}

	for _, view := range cfg.Views {
		var includesRee []*regexp.Regexp
		var excludesRee []*regexp.Regexp
		if view.Pattern != "" {
			view.Includes = append(view.Includes, view.Pattern)
		}
		for _, pattern := range view.Includes {
			ree, _ := regexp.Compile(makePattern(pattern))
			includesRee = append(includesRee, ree)
		}
		for _, pattern := range view.Excludes {
			ree, _ := regexp.Compile(makePattern(pattern))
			excludesRee = append(excludesRee, ree)
		}
		views = append(views, ViewTemplate{view, includesRee, excludesRee})
	}
	return views
}

func loadState(state *servicesState, be backend.Backend, cfg config.Config) {
	state.viewTemplates = loadViewTemplates(cfg)
	state.viewStates = be.LoadViews()

	for _, data := range be.LoadServices() {
		service := &Service{data: *data}
		service.updateExpiry()
		state.services[data.Name] = service

		addViewsToService(state, service, []stateUpdate{})
	}
	return
}
