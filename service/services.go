package service

import (
	"github.com/boivie/lovebeat/backend"
	"github.com/boivie/lovebeat/config"
	"github.com/boivie/lovebeat/eventbus"
	"github.com/boivie/lovebeat/model"
	"github.com/op/go-logging"

	"github.com/boivie/lovebeat/alert"
	"github.com/boivie/lovebeat/notify"
	"time"
)

type ServicesImpl struct {
	updateChan      chan *Update
	getServicesChan chan *getServicesCmd
	getServiceChan  chan *getServiceCmd
	getViewsChan    chan *getViewsCmd
	getViewChan     chan *getViewCmd
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

func printUpdates(updates []stateUpdate) {
	for _, update := range updates {
		if update.newService != nil && update.oldService == nil {
			log.Infof("SERVICE '%s', created -> %s", update.newService.data.Name, update.newService.data.State)
		} else if update.newService == nil && update.oldService != nil {
			log.Infof("SERVICE '%s', %s -> deleted", update.oldService.data.Name, update.oldService.data.State)
		} else if update.newService != nil && update.oldService != nil {
			if update.newService.data.State != update.oldService.data.State {
				log.Infof("SERVICE '%s', state %s -> %s", update.oldService.data.Name, update.oldService.data.State, update.newService.data.State)
			}
			if update.newService.data.Timeout != update.oldService.data.Timeout {
				log.Infof("SERVICE '%s', tmo %d -> %d", update.oldService.data.Name, update.oldService.data.Timeout, update.newService.data.Timeout)
			}
		} else if update.newView != nil && update.oldView == nil {
			log.Infof("VIEW '%s', created -> %s", update.newView.data.Name, update.newView.data.State)
		} else if update.newView != nil && update.oldView != nil {
			log.Infof("VIEW '%s', state %s -> %s", update.oldView.data.Name, update.oldView.data.State, update.newView.data.State)
		}
	}
}

func sendBusEvents(bus *eventbus.EventBus, updates []stateUpdate) {
	for _, update := range updates {
		if update.newService != nil && update.oldService == nil {
			bus.Publish(model.ServiceAddedEvent{Service: update.newService.getExternalModel()})
		} else if update.newService == nil && update.oldService != nil {
			bus.Publish(model.ServiceRemovedEvent{Service: update.oldService.getExternalModel()})
		} else if update.newService != nil && update.oldService != nil {
			bus.Publish(model.ServiceStateChangedEvent{
				Service:  update.newService.getExternalModel(),
				Previous: update.oldService.data.State,
				Current:  update.newService.data.State,
			})
		} else if update.newView != nil && update.oldView == nil {
			bus.Publish(model.ViewAddedEvent{update.newView.getExternalModel()})
		} else if update.newView != nil && update.oldView != nil {
			bus.Publish(model.ViewStateChangedEvent{
				View:     update.newView.getExternalModel(),
				Previous: update.oldView.data.State,
				Current:  update.newView.data.State,
			})
		} else if update.newView == nil && update.oldView != nil {
			bus.Publish(model.ViewRemovedEvent{update.oldView.getExternalModel()})
		}
	}
}

func triggerAlerters(alerter alert.Alerter, updates []stateUpdate) {
	for _, update := range updates {
		if update.newView != nil {
			oldState := model.StateNew
			if update.oldView != nil {
				oldState = update.oldView.data.State
			}
			alerter.Notify(alert.AlertInfo{
				View:       update.newView.getExternalModel(),
				Previous:   oldState,
				Current:    update.newView.data.State,
				ViewConfig: update.newView.tmpl.config,
			})
		}
	}
}

func NewServices(beiface backend.Backend, bus *eventbus.EventBus, alerter alert.Alerter, cfg config.Config, notifier notify.Notifier) Services {
	svcs := ServicesImpl{
		updateChan:      make(chan *Update, MAX_UNPROCESSED_PACKETS),
		getServicesChan: make(chan *getServicesCmd, 5),
		getServiceChan:  make(chan *getServiceCmd, 5),
		getViewsChan:    make(chan *getViewsCmd, 5),
		getViewChan:     make(chan *getViewCmd, 5),
	}

	go svcs.Monitor(cfg, notifier, beiface, bus, alerter)

	go func() {
		ticker := time.NewTicker(time.Duration(1) * time.Second)
		for _ = range ticker.C {
			svcs.updateChan <- &Update{Ts: tsnow(), Tick: &Tick{}}
		}
	}()

	return &svcs
}
