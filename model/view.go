package model

const (
	STATE_PAUSED  = "paused"
	STATE_OK      = "ok"
	STATE_WARNING = "warning"
	STATE_ERROR   = "error"
)

const PREVIOUS_BEATS_COUNT = 20

type Service struct {
	Name           string
	LastValue      int
	LastBeat       int64
	PreviousBeats  []int64
	LastUpdated    int64
	WarningTimeout int64
	ErrorTimeout   int64
	State          string
}

type View struct {
	Name        string
	State       string
	Regexp      string
	LastUpdated int64
	IncidentNbr int
	Alerts      []string
}
