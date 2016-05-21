package alert

import (
	"github.com/boivie/lovebeat/config"
	"github.com/boivie/lovebeat/model"
	"github.com/boivie/lovebeat/notify"
	"github.com/boivie/lovebeat/service"
	"github.com/op/go-logging"
	"time"
)

var log = logging.MustGetLogger("lovebeat")

type alerter struct {
	q chan AlertInfo
}

func (s *alerter) OnUpdate(ts int64, update model.Update) {}

func (s *alerter) OnServiceAdded(ts int64, service model.Service)                  {}
func (s *alerter) OnServiceUpdated(ts int64, oldService, newService model.Service) {}
func (s *alerter) OnServiceRemoved(ts int64, service model.Service)                {}

func (s *alerter) OnViewAdded(ts int64, view model.View, config config.ConfigView) {}

func (s *alerter) OnViewUpdated(ts int64, oldView, newView model.View, config config.ConfigView) {
	if oldView.State != newView.State {
		s.q <- AlertInfo{
			View:       newView,
			Previous:   oldView.State,
			Current:    newView.State,
			ViewConfig: config,
		}
	}
}

func (s *alerter) OnViewRemoved(ts int64, view model.View, config config.ConfigView) {
	// TODO: Send alerts?
}

type AlertInfo struct {
	View       model.View
	Previous   string
	Current    string
	ViewConfig config.ConfigView
}

type AlerterBackend interface {
	Notify(alertCfg config.ConfigAlert, ev AlertInfo)
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

func New(cfg config.Config, notifier notify.Notifier) service.ServiceCallback {
	var q = make(chan AlertInfo, 1000)
	go runner(cfg, q, notifier)
	return &alerter{q}
}
