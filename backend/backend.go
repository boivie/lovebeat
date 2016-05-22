package backend

import (
	"github.com/boivie/lovebeat/model"
)

type Backend interface {
	LoadServices() []*model.Service
	LoadAlarms() []*model.Alarm

	SaveService(service *model.Service)
	SaveAlarm(alarm *model.Alarm)

	DeleteService(name string)
	DeleteAlarm(name string)
	Sync()
}
