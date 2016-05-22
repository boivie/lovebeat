package model

// The service is new (this is an initial state from where it will transition quickly)
const StateNew = "new"

// The service is perfectly fine
const StateOk = "ok"

// The service is in an error state
const StateError = "error"

// The service is muted
const StateMuted = "muted"

// The number of beats that we save
const BeatHistoryCount = 100

// Special values for error and warning timeouts
const TIMEOUT_CLEAR int64 = -1
const TIMEOUT_AUTO int64 = -2

// Service is something that can issue a beat
type Service struct {
	Name            string   `json:"name"`                   // Name of the service
	LastBeat        int64    `json:"last_beat"`              // Timestamp, in milliseconds since epoch, of last beat
	BeatHistory     []int64  `json:"beat_history,omitempty"` // The last X duration (in milliseconds) between heartbeats
	Timeout         int64    `json:"timeout"`                // The timeout, in milliseconds
	State           string   `json:"state"`                  // One of the StateXX constants
	MutedSince      int64    `json:"muted_since"`            // Since when this service has been muted (if >0)
	LastStateChange int64    `json:"last_state_change"`      // Time when the service last changed state
	InAlarms        []string `json:"in_alarms,omitempty"`    // A list of alarms this service is included in
}

// Alarm contains a number of services and trigger alerts when it's not OK.
type Alarm struct {
	Name            string   `json:"name"`              // Name of the alarm
	State           string   `json:"state"`             // One of the StateXX constant
	IncidentNbr     int      `json:"incident_nbr"`      // Incrementing number everytime the alarm leaves the StateOk state
	FailedServices  []string `json:"failed_services"`   // A list of failed services, if any. This list may be trimmed.
	LastStateChange int64    `json:"last_state_change"` // Time when the alarm last changed state
}
