package alert

import (
	"github.com/boivie/lovebeat-go/backend"
	"github.com/op/go-logging"
)

var (
	log = logging.MustGetLogger("lovebeat")
)

type Alerter interface {
	Notify(previous backend.StoredView,
		current backend.StoredView,
		servicesInError []backend.StoredService)
}
