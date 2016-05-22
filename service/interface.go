package service

import (
	"github.com/boivie/lovebeat/config"
	"github.com/boivie/lovebeat/model"
)

type Services interface {
	Subscribe(cb ServiceCallback)
	Update(update *model.Update)

	GetServices() []model.Service
	GetServicesInAlarm(alarm string) []model.Service
	GetService(name string) *model.Service
	GetAlarms() []model.Alarm
	GetAlarm(name string) *model.Alarm
}

type ServiceCallback interface {
	OnUpdate(ts int64, update model.Update)

	OnServiceAdded(ts int64, service model.Service)
	OnServiceUpdated(ts int64, oldService, newService model.Service)
	OnServiceRemoved(ts int64, service model.Service)

	OnAlarmAdded(ts int64, alarm model.Alarm, config config.ConfigAlarm)
	OnAlarmUpdated(ts int64, oldAlarm, newAlarm model.Alarm, config config.ConfigAlarm)
	OnAlarmRemoved(ts int64, alarm model.Alarm, config config.ConfigAlarm)
}
