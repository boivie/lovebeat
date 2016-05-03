package alert

import (
	"github.com/boivie/lovebeat/config"
	"github.com/boivie/lovebeat/model"
	"github.com/boivie/lovebeat/notify"
	"github.com/op/go-logging"
	"time"
)

var log = logging.MustGetLogger("lovebeat")

type AlertInfo struct {
	View           model.View
	Previous       string
	Current        string
	FailedServices []model.Service
	ViewConfig     config.ConfigView
}

type AlerterBackend interface {
	Notify(alertCfg config.ConfigAlert, ev AlertInfo)
}

type Alerter interface {
	Notify(ev AlertInfo)
}

type alerter struct {
	q chan AlertInfo
}

func (f alerter) Notify(ev AlertInfo) {
	if ev.Current != model.StatePaused && ev.Previous != model.StatePaused {
		f.q <- ev
	}
}

func runner(cfg config.Config, q <-chan AlertInfo, notifier notify.Notifier) {
	alerters := []AlerterBackend{
		NewMailAlerter(cfg),
		NewSlackAlerter(cfg),
		NewWebhooksAlerter(cfg),
		NewScriptAlerter(cfg),
	}

	healthCheck := time.NewTicker(time.Duration(60) * time.Second)
	for {
		select {
		case <-healthCheck.C:
			notifier.Notify("alerter")
		case event := <-q:
			for _, alert := range event.ViewConfig.Alerts {
				if alertCfg, ok := cfg.Alerts[alert]; ok {
					for _, a := range alerters {
						a.Notify(alertCfg, event)
					}
				}
			}
		}
	}
}

func Init(cfg config.Config, notifier notify.Notifier) Alerter {
	var q = make(chan AlertInfo, 1000)
	go runner(cfg, q, notifier)
	return &alerter{q}
}
