package service

import (
	"github.com/boivie/lovebeat/backend"
	"github.com/boivie/lovebeat/config"
	"regexp"
)

func loadViewTemplates(cfg config.Config) []ViewTemplate {
	views := make([]ViewTemplate, 0)

	views = append(views, ViewTemplate{config.ConfigView{Name: "all"}, regexp.MustCompile("")})

	for _, view := range cfg.Views {
		ree, _ := regexp.Compile(makePattern(view.Pattern))
		views = append(views, ViewTemplate{view, ree})
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
