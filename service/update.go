package service

import (
	"github.com/boivie/lovebeat/model"
)

func addAlarmsToService(state *servicesState, svc *service, prevUpdates []stateUpdate) (updates []stateUpdate) {
	updates = prevUpdates
	for _, tmpl := range state.alarmTemplates {
		name := tmpl.makeName(svc.data.Name)
		if name != "" {
			a, exists := state.alarms[name]
			if !exists {
				a = &alarm{tmpl: tmpl, data: model.Alarm{
					Name:        name,
					State:       model.StateNew,
					IncidentNbr: 0,
				}}
				for _, existingAlarm := range state.alarmStates {
					if existingAlarm.Name == name {
						a.data = *existingAlarm
					}
				}
				state.alarms[name] = a
				alarmCopy := *a
				updates = append(updates, stateUpdate{oldAlarm: nil, newAlarm: &alarmCopy})
			}
			a.servicesInAlarm = append(a.servicesInAlarm, svc)
			svc.inAlarms = append(svc.inAlarms, a)
		}
	}
	return
}

func removeAlarmsFromService(state *servicesState, svc *service, prevUpdates []stateUpdate) (updates []stateUpdate) {
	updates = prevUpdates
	for _, alarm := range svc.inAlarms {
		var remainingServices []*service
		for _, s := range alarm.servicesInAlarm {
			if s != svc {
				remainingServices = append(remainingServices, s)
			}
		}
		alarm.servicesInAlarm = remainingServices
	}
	return
}

func updateServices(state *servicesState, cmd *model.Update) (updates []stateUpdate) {
	if cmd.Tick != nil {
		for _, service := range state.services {
			if service.data.State != service.stateAt(cmd.Ts) {
				var ref = *service
				service.data.State = service.stateAt(cmd.Ts)
				service.data.LastStateChange = cmd.Ts
				updates = append(updates, stateUpdate{oldService: &ref, newService: service})
			}
		}
	} else if cmd.Service != "" {
		updated := false
		var old *service
		service := state.services[cmd.Service]
		if service == nil {
			service = newService(cmd.Service)
			state.services[cmd.Service] = service
			updates = addAlarmsToService(state, service, updates)
			updated = true
		} else {
			var r = *service
			old = &r
		}

		if cmd.Beat != nil {
			service.registerBeat(cmd.Ts)
			updated = true
		}
		if cmd.SetTimeout != nil && service.data.Timeout != cmd.SetTimeout.Timeout {
			service.data.Timeout = cmd.SetTimeout.Timeout
			updated = true
		}
		if cmd.MuteService != nil {
			if cmd.MuteService.Muted && service.data.MutedSince == 0 {
				service.data.MutedSince = cmd.Ts
				updated = true
			} else if !cmd.MuteService.Muted && service.data.MutedSince != 0 {
				service.data.MutedSince = 0
				updated = true
			}
		}
		if cmd.DeleteService != nil {
			updates = removeAlarmsFromService(state, service, updates)
			delete(state.services, service.name())
			service = nil
			updated = old != nil
		}

		if service != nil {
			service.updateExpiry()
			state := service.stateAt(cmd.Ts)
			if state != service.data.State {
				service.data.State = state
				service.data.LastStateChange = cmd.Ts
			}
		}

		if updated {
			updates = append(updates, stateUpdate{oldService: old, newService: service})
		}
	}
	return
}

func removeAlarms(state *servicesState, c *model.Update, prevUpdates []stateUpdate) (updates []stateUpdate) {
	updates = prevUpdates
	if c.DeleteAlarm != nil {
		alarm, exist := state.alarms[c.Alarm]
		if exist && len(alarm.servicesInAlarm) == 0 {
			delete(state.alarms, c.Alarm)
			updates = append(updates, stateUpdate{oldAlarm: alarm, newAlarm: nil})
		}
	}
	return
}

func updateAlarms(state *servicesState, ts int64, prevUpdates []stateUpdate) (updates []stateUpdate) {
	updates = prevUpdates
	for _, update := range prevUpdates {
		service := update.newService
		if service == nil {
			service = update.oldService
		}
		if service != nil {
			for _, alarm := range service.inAlarms {
				newState := alarm.calculateState()
				if alarm.data.State != newState {
					ref := *alarm
					alarm.data.State = newState
					alarm.data.LastStateChange = ts
					if ref.data.State == model.StateOk || ref.data.State == model.StateNew {
						alarm.data.IncidentNbr += 1
					}
					updates = append(updates, stateUpdate{oldAlarm: &ref, newAlarm: alarm})
				}
			}
		}
	}
	return
}
