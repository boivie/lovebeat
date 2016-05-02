package model

// The service is paused and will not trigger any alarms
const StatePaused = "paused"

// The service is perfectly fine
const StateOk = "ok"

// The service is in an error state
const StateError = "error"

// The number of beats that we save
const BeatHistoryCount = 100

// Service is something that can issue a beat
type Service struct {
	Name        string  `json:"name"`         // Name of the service
	LastBeat    int64   `json:"last_beat"`    // Timestamp, in milliseconds since epoch, of last beat
	BeatHistory []int64 `json:"beat_history"` // The last X duration (in milliseconds) between heartbeats
	Timeout     int64   `json:"timeout"`      // The timeout, in milliseconds
	State       string  `json:"state"`        // One of the StateXX constants
}

// View is a collection of services
type View struct {
	Name        string   `json:"name"`         // Name of the view
	State       string   `json:"state"`        // One of the StateXX constant
	IncidentNbr int      `json:"incident_nbr"` // Incrementing number everytime the view leaves the StateOk state
}
