package model

// The service is paused and will not trigger any alarms
const StatePaused = "paused"

// The service is perfectly fine
const StateOk = "ok"

// The service is in a warning state
const StateWarning = "warning"

// The service is in an error state
const StateError = "error"

// The number of beats that we save
const BeatHistoryCount = 100

// Service is something that can issue a beat
type Service struct {
	Name           string  // Name of the service
	LastBeat       int64   // Timestamp, in milliseconds since epoch, of last beat
	BeatHistory    []int64 // The last X duration (in milliseconds) between heartbeats
	WarningTimeout int64   // The warning timeout, in milliseconds
	ErrorTimeout   int64   // The error timeout, in milliseconds
	State          string  // One of the StateXX constants
}

// View is a collection of services
type View struct {
	Name        string // Name of the view
	State       string // One of the StateXX constant
	Regexp      string // Services matching this expression will be included in the view
	IncidentNbr int    // Incrementing number everytime the view leaves the StateOk state
	Alerts      []string
}
