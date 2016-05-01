package alert

import (
	"github.com/boivie/lovebeat/config"
	"github.com/boivie/lovebeat/eventbus"
	"github.com/boivie/lovebeat/model"
	"github.com/boivie/lovebeat/service"
	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("lovebeat")

type Alerter interface {
	Notify(cfg config.ConfigAlert, ev service.ViewStateChangedEvent)
}


func RegisterAlerters(bus *eventbus.EventBus, cfg config.Config, client service.ServiceIf) {
	alerters := []Alerter{
		NewMailAlerter(cfg),
		NewSlackAlerter(cfg),
		NewWebhooksAlerter(cfg),
		NewScriptAlerter(cfg),
	}

	var q = make(chan service.ViewStateChangedEvent, 100)
	go func() {
		for ev := range q {
			for _, alert := range client.GetViewAlerts(ev.View.Name) {
				for _, alerter := range alerters {
					if cfg, ok := cfg.Alerts[alert]; ok {
						alerter.Notify(cfg, ev)
					}
				}
			}
		}
	}()

	bus.RegisterHandler(func(ev service.ViewStateChangedEvent) {
		if ev.Current != model.StatePaused && ev.Previous != model.StatePaused {
			q <- ev
		}
	})
}
