package backend

const (
	STATE_PAUSED  = "paused"
	STATE_OK      = "ok"
	STATE_WARNING = "warning"
	STATE_ERROR   = "error"
)

const PREVIOUS_BEATS_COUNT = 20

type StoredService struct {
	Name           string
	LastValue      int
	LastBeat       int64
	PreviousBeats  []int64
	LastUpdated    int64
	WarningTimeout int64
	ErrorTimeout   int64
	State          string
}

type StoredView struct {
	Name        string
	State       string
	Regexp      string
	LastUpdated int64
	AlertMail   string
}

type Backend interface {
	LoadServices() []*StoredService
	LoadViews() []*StoredView

	SaveService(service *StoredService)
	SaveView(view *StoredView)

	DeleteService(name string)
	DeleteView(name string)
}
