package service

import (
	"github.com/boivie/lovebeat/model"
)

func addViewsToService(state *servicesState, svc *Service, prevUpdates []stateUpdate) (updates []stateUpdate) {
	updates = prevUpdates
	for _, tmpl := range state.viewTemplates {
		name := tmpl.makeName(svc.data.Name)
		if name != "" {
			view, exists := state.views[name]
			if !exists {
				view = &View{tmpl: &tmpl, data: model.View{
					Name:        name,
					State:       model.StatePaused,
					IncidentNbr: 0,
				}}
				for _, existingView := range state.viewStates {
					if existingView.Name == name {
						view.data = *existingView
					}
				}
				state.views[name] = view
				updates = append(updates, stateUpdate{oldView: nil, newView: view})
			}
			view.servicesInView = append(view.servicesInView, svc)
			svc.inViews = append(svc.inViews, view)
		}
	}
	return
}

func updateServices(state *servicesState, cmd *Update) (updates []stateUpdate) {
	if cmd.Tick != nil {
		for _, service := range state.services {
			if service.data.State != model.StatePaused && service.data.State != service.stateAt(cmd.Ts) {
				var ref = *service
				service.data.State = service.stateAt(cmd.Ts)
				service.data.LastStateChange = cmd.Ts
				updates = append(updates, stateUpdate{oldService: &ref, newService: service})
			}
		}
	} else {
		updated := false
		var old *Service
		service := state.services[cmd.Service]
		if service == nil {
			service = newService(cmd.Service)
			state.services[cmd.Service] = service
			updates = addViewsToService(state, service, updates)
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

func updateViews(state *servicesState, ts int64, prevUpdates []stateUpdate) (updates []stateUpdate) {
	updates = prevUpdates
	for _, update := range updates {
		service := update.newService
		if service == nil {
			service = update.oldService
		}
		if service != nil {
			for _, view := range service.inViews {
				newState := view.calculateState()
				if view.data.State != newState {
					ref := *view
					view.data.State = newState
					view.data.LastStateChange = ts
					if ref.data.State == model.StateOk {
						view.data.IncidentNbr += 1
					}
					updates = append(updates, stateUpdate{oldView: &ref, newView: view})
				}
			}
		}
	}
	return
}
