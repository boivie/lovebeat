package alert

import (
	"github.com/boivie/lovebeat/backend"
	"github.com/op/go-logging"
)

var (
	log = logging.MustGetLogger("lovebeat")
)

type Alert struct {
	Previous        backend.StoredView
	Current         backend.StoredView
	ServicesInError []backend.StoredService
}

type Alerter interface {
	Notify(alert Alert)
}
