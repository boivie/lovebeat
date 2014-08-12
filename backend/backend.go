package backend

const (
	STATE_PAUSED  = "paused"
	STATE_OK      = "ok"
	STATE_WARNING = "warning"
	STATE_ERROR   = "error"
)

type StoredService struct {
	Name           string
	LastValue      int
	LastBeat       int64
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
}

type Backend interface {
	LoadServices() []*StoredService
	LoadViews() []*StoredView

	AddServiceLog(name string, ts int64, action string, extra string)
	AddViewLog(name string, ts int64, action string, extra string)

	SaveService(service *StoredService)
	SaveView(view *StoredView)
}
