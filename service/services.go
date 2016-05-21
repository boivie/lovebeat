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
	getServicesChan chan *getServicesCmd
	getServiceChan  chan *getServiceCmd
	getViewsChan    chan *getViewsCmd
	getViewChan     chan *getViewCmd
	subscribeChan   chan ServiceCallback
}

const (
	MAX_UNPROCESSED_PACKETS = 1000
)

var log = logging.MustGetLogger("lovebeat")

type servicesState struct {
	viewTemplates []ViewTemplate
	viewStates    []*model.View
	services      map[string]*Service
	views         map[string]*View
}

type stateUpdate struct {
	oldService *Service
	newService *Service
	oldView    *View
	newView    *View
}

func newState() *servicesState {
	return &servicesState{

		services: make(map[string]*Service),
		views:    make(map[string]*View),
	}
}

func persist(be backend.Backend, updates []stateUpdate) {
	for _, u := range updates {
		if u.newService != nil {
			be.SaveService(&u.newService.data)
		} else if u.oldService != nil {
			be.DeleteService(u.oldService.data.Name)
		} else if u.newView != nil {
			be.SaveView(&u.newView.data)
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
			} else if update.newView != nil && update.oldView == nil {
				observer.OnViewAdded(c.Ts, update.newView.getExternalModel(), update.newView.tmpl.config)
			} else if update.newView != nil && update.oldView != nil {
				observer.OnViewUpdated(c.Ts, update.oldView.getExternalModel(), update.newView.getExternalModel(), update.newView.tmpl.config)
			} else if update.newView == nil && update.oldView != nil {
				observer.OnViewRemoved(c.Ts, update.oldView.getExternalModel(), update.oldView.tmpl.config)
			}
		}
	}
}

func NewServices(beiface backend.Backend, cfg config.Config, notifier notify.Notifier) Services {
	svcs := ServicesImpl{
		subscribeChan:   make(chan ServiceCallback, 10),
		updateChan:      make(chan *model.Update, MAX_UNPROCESSED_PACKETS),
		getServicesChan: make(chan *getServicesCmd, 5),
		getServiceChan:  make(chan *getServiceCmd, 5),
		getViewsChan:    make(chan *getViewsCmd, 5),
		getViewChan:     make(chan *getViewCmd, 5),
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
