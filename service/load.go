package service

import (
	"github.com/boivie/lovebeat/backend"
	"github.com/boivie/lovebeat/config"
	"regexp"
)

func loadAlarmTemplates(cfg config.Config) []alarmTemplate {
	alarms := make([]alarmTemplate, 0)

	for _, alarm := range cfg.Alarms {
		var includesRee []*regexp.Regexp
		var excludesRee []*regexp.Regexp
		if alarm.Pattern != "" {
			alarm.Includes = append(alarm.Includes, alarm.Pattern)
		}
		for _, pattern := range alarm.Includes {
			ree, _ := regexp.Compile(makePattern(pattern))
			includesRee = append(includesRee, ree)
		}
		for _, pattern := range alarm.Excludes {
			ree, _ := regexp.Compile(makePattern(pattern))
			excludesRee = append(excludesRee, ree)
		}
		alarms = append(alarms, alarmTemplate{alarm, includesRee, excludesRee})
	}
	return alarms
}

func loadState(state *servicesState, be backend.Backend, cfg config.Config) {
	state.alarmTemplates = loadAlarmTemplates(cfg)
	state.alarmStates = be.LoadAlarms()

	for _, data := range be.LoadServices() {
		service := &service{data: *data}
		service.updateExpiry()
		state.services[data.Name] = service

		addAlarmsToService(state, service, []stateUpdate{})
	}
	return
}
