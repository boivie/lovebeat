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
	Name           string
	LastValue      int
	LastBeat       int64
	BeatHistory    []int64
	WarningTimeout int64
	ErrorTimeout   int64
	State          string
}

// View is a collection of services
type View struct {
	Name        string
	State       string
	Regexp      string
	IncidentNbr int
	Alerts      []string
}
