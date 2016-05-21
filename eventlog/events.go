package eventlog

import "github.com/boivie/lovebeat/model"

type UpdateEvent struct {
	Update *model.Update `json:"update"`
}

// When adding a new expression struct type here, don't forget
// to add it to the test cases so the member names are checked
// for conformity.

// ServiceAddedEvent will be sent when a service has been added
type ServiceAddedEvent struct {
	Service model.Service `json:"service"`
}

// ServiceStateChangedEvent will be sent when a service has changed state
type ServiceStateChangedEvent struct {
	Service  model.Service `json:"service"`
	Previous string        `json:"previous"`
	Current  string        `json:"current"`
}

// ServiceRemovedEvent will be sent when a service has been removed
type ServiceRemovedEvent struct {
	Service model.Service `json:"service"`
}

// ViewAddedEvent will be sent when a view has been added
type ViewAddedEvent struct {
	View model.View `json:"view"`
}

// ViewStateChangedEvent will be sent when a view has changed state
type ViewStateChangedEvent struct {
	View     model.View `json:"view"`
	Previous string     `json:"previous"`
	Current  string     `json:"current"`
}

// ViewRemovedEvent will be sent when a view has been manually removed. It will not be sent
// if a view has been removed by the configuration as loaded on startup
type ViewRemovedEvent struct {
	View model.View `json:"view"`
}

type Event struct {
	Ts   int64  `json:"ts"`
	Type string `json:"type"`

	ServiceAddedEvent        *ServiceAddedEvent        `json:"service_added,omitempty"`
	ServiceStateChangedEvent *ServiceStateChangedEvent `json:"service_state_changed,omitempty"`
	ServiceRemovedEvent      *ServiceRemovedEvent      `json:"service_removed,omitempty"`

	ViewAddedEvent        *ViewAddedEvent        `json:"view_added,omitempty"`
	ViewStateChangedEvent *ViewStateChangedEvent `json:"view_state_changed,omitempty"`
	ViewRemovedEvent      *ViewRemovedEvent      `json:"view_removed,omitempty"`
}
