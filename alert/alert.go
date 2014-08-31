package alert

import (
	"github.com/boivie/lovebeat-go/backend"
)

type Alerter interface {
	Notify(previous backend.StoredView,
		current backend.StoredView,
		servicesInError []backend.StoredService)
}
