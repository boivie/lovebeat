package alert

import (
	"github.com/boivie/lovebeat/model"
	"github.com/op/go-logging"
)

var (
	log = logging.MustGetLogger("lovebeat")
)

type Alert struct {
	Previous        model.View
	Current         model.View
	ServicesInError []model.Service
}

type Alerter interface {
	Notify(alert Alert)
}
