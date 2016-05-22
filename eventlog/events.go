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

// AlarmAddedEvent will be sent when an alarm has been added
type AlarmAddedEvent struct {
	Alarm model.Alarm `json:"alarm"`
}

// AlarmStateChangedEvent will be sent when an alarm has changed state
type AlarmStateChangedEvent struct {
	Alarm    model.Alarm `json:"alarm"`
	Previous string      `json:"previous"`
	Current  string      `json:"current"`
}

// AlarmRemovedEvent will be sent when an alarm has been manually removed. It will not be sent
// if an alarm has been removed by the configuration as loaded on startup
type AlarmRemovedEvent struct {
	Alarm model.Alarm `json:"alarm"`
}

type Event struct {
	Ts   int64  `json:"ts"`
	Type string `json:"type"`

	ServiceAddedEvent        *ServiceAddedEvent        `json:"service_added,omitempty"`
	ServiceStateChangedEvent *ServiceStateChangedEvent `json:"service_state_changed,omitempty"`
	ServiceRemovedEvent      *ServiceRemovedEvent      `json:"service_removed,omitempty"`

	AlarmAddedEvent   *AlarmAddedEvent        `json:"alarm_added,omitempty"`
	AlarmStateChanged *AlarmStateChangedEvent `json:"alarm_state_changed,omitempty"`
	AlarmRemovedEvent *AlarmRemovedEvent      `json:"alarm_removed,omitempty"`
}
