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

func (s *alerter) OnAlarmAdded(ts int64, alarm model.Alarm, config config.ConfigAlarm) {}

func (s *alerter) OnAlarmUpdated(ts int64, oldAlarm, newAlarm model.Alarm, config config.ConfigAlarm) {
	if oldAlarm.State != newAlarm.State {
		s.q <- AlertInfo{
			Alarm:       newAlarm,
			Previous:    oldAlarm.State,
			Current:     newAlarm.State,
			AlarmConfig: config,
		}
	}
}

func (s *alerter) OnAlarmRemoved(ts int64, alarm model.Alarm, config config.ConfigAlarm) {
	// TODO: Send alerts?
}

type AlertInfo struct {
	Alarm       model.Alarm
	Previous    string
	Current     string
	AlarmConfig config.ConfigAlarm
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
			for _, alert := range event.AlarmConfig.Alerts {
				for _, a := range alerters {
					a.Notify(alert, event)
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
