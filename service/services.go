package service

import (
	"github.com/boivie/lovebeat/backend"
	"github.com/boivie/lovebeat/config"
	"github.com/boivie/lovebeat/model"
	"github.com/op/go-logging"

	"github.com/boivie/lovebeat/notify"
	"time"
)

type ServicesImpl struct {
	updateChan      chan *model.Update
	getServicesChan chan *getServicesInAlarmCmd
	getServiceChan  chan *getServiceCmd
	getAlarmsChan   chan *getAlarmsCmd
	getAlarmChan    chan *getAlarmCmd
	subscribeChan   chan ServiceCallback
}

const (
	MAX_UNPROCESSED_PACKETS = 1000
)

var log = logging.MustGetLogger("lovebeat")

type servicesState struct {
	alarmTemplates []alarmTemplate
	alarmStates    []*model.Alarm
	services       map[string]*service
	alarms         map[string]*alarm
}

type stateUpdate struct {
	oldService *service
	newService *service
	oldAlarm   *alarm
	newAlarm   *alarm
}

func newState() *servicesState {
	return &servicesState{

		services: make(map[string]*service),
		alarms:   make(map[string]*alarm),
	}
}

func persist(be backend.Backend, updates []stateUpdate) {
	for _, u := range updates {
		if u.newService != nil {
			be.SaveService(&u.newService.data)
		} else if u.oldService != nil {
			be.DeleteService(u.oldService.data.Name)
		} else if u.newAlarm != nil {
			be.SaveAlarm(&u.newAlarm.data)
		}
	}
}

func sendCallbacks(observers []ServiceCallback, c *model.Update, updates []stateUpdate) {
	for _, observer := range observers {
		observer.OnUpdate(c.Ts, *c)
		for _, update := range updates {
			if update.newService != nil && update.oldService == nil {
				observer.OnServiceAdded(c.Ts, update.newService.getExternalModel())
			} else if update.newService == nil && update.oldService != nil {
				observer.OnServiceRemoved(c.Ts, update.oldService.getExternalModel())
			} else if update.newService != nil && update.oldService != nil {
				observer.OnServiceUpdated(c.Ts, update.oldService.getExternalModel(), update.newService.getExternalModel())
			} else if update.newAlarm != nil && update.oldAlarm == nil {
				observer.OnAlarmAdded(c.Ts, update.newAlarm.getExternalModel(), update.newAlarm.tmpl.config)
			} else if update.newAlarm != nil && update.oldAlarm != nil {
				observer.OnAlarmUpdated(c.Ts, update.oldAlarm.getExternalModel(), update.newAlarm.getExternalModel(), update.newAlarm.tmpl.config)
			} else if update.newAlarm == nil && update.oldAlarm != nil {
				observer.OnAlarmRemoved(c.Ts, update.oldAlarm.getExternalModel(), update.oldAlarm.tmpl.config)
			}
		}
	}
}

func NewServices(beiface backend.Backend, cfg config.Config, notifier notify.Notifier) Services {
	svcs := ServicesImpl{
		subscribeChan:   make(chan ServiceCallback, 10),
		updateChan:      make(chan *model.Update, MAX_UNPROCESSED_PACKETS),
		getServicesChan: make(chan *getServicesInAlarmCmd, 5),
		getServiceChan:  make(chan *getServiceCmd, 5),
		getAlarmsChan:   make(chan *getAlarmsCmd, 5),
		getAlarmChan:    make(chan *getAlarmCmd, 5),
	}

	go svcs.Monitor(cfg, notifier, beiface)

	go func() {
		ticker := time.NewTicker(time.Duration(1) * time.Second)
		for _ = range ticker.C {
			svcs.updateChan <- &model.Update{Ts: tsnow(), Tick: &model.Tick{}}
		}
	}()

	return &svcs
}
