package alert

import (
	"github.com/boivie/lovebeat/config"
	"github.com/boivie/lovebeat/eventbus"
	"github.com/boivie/lovebeat/service"
	"github.com/op/go-logging"
)

var (
	log      = logging.MustGetLogger("lovebeat")
	alerters []Alerter
	alerts   map[string]config.ConfigAlert
)

type Alerter interface {
	Notify(cfg config.ConfigAlert, ev service.ViewStateChangedEvent)
}

func ViewStateChanged(ev service.ViewStateChangedEvent) {
	for _, alert := range ev.View.Alerts {
		for _, alerter := range alerters {
			if cfg, ok := alerts[alert]; ok {
				alerter.Notify(cfg, ev)
			}
		}
	}
}

func RegisterAlerters(bus *eventbus.EventBus, cfg config.Config) {
	alerters = []Alerter{
		NewMailAlerter(cfg),
		NewWebhooksAlerter(cfg),
	}
	alerts = cfg.Alerts
	bus.RegisterHandler(ViewStateChanged)
}
