package eventlog

import (
	"encoding/json"
	"github.com/boivie/lovebeat/config"
	"github.com/boivie/lovebeat/model"
	"github.com/boivie/lovebeat/service"
	"github.com/op/go-logging"
	"io"
	"io/ioutil"
)

var log = logging.MustGetLogger("lovebeat")

type eventLog struct {
	writer io.Writer
}

func (s *eventLog) OnUpdate(ts int64, update model.Update) {}

func (s *eventLog) OnServiceAdded(ts int64, service model.Service) {
	s.log(Event{Ts: ts, Type: "service_added",
		ServiceAddedEvent: &ServiceAddedEvent{service}})
}
func (s *eventLog) OnServiceUpdated(ts int64, oldService, newService model.Service) {
	if oldService.State != newService.State {
		newService.BeatHistory = nil
		newService.InViews = nil
		s.log(Event{Ts: ts, Type: "service_state_changed",
			ServiceStateChangedEvent: &ServiceStateChangedEvent{newService, oldService.State, newService.State}})
	}
}
func (s *eventLog) OnServiceRemoved(ts int64, service model.Service) {
	s.log(Event{Ts: ts, Type: "service_removed",
		ServiceRemovedEvent: &ServiceRemovedEvent{service}})
}

func (s *eventLog) OnViewAdded(ts int64, view model.View, config config.ConfigView) {
	s.log(Event{Ts: ts, Type: "view_added",
		ViewAddedEvent: &ViewAddedEvent{view}})
}
func (s *eventLog) OnViewUpdated(ts int64, oldView, newView model.View, config config.ConfigView) {
	if oldView.State != newView.State {
		s.log(Event{Ts: ts, Type: "view_state_changed",
			ViewStateChangedEvent: &ViewStateChangedEvent{newView, oldView.State, newView.State}})
	}
}
func (s *eventLog) OnViewRemoved(ts int64, view model.View, config config.ConfigView) {
	s.log(Event{Ts: ts, Type: "view_removed",
		ViewRemovedEvent: &ViewRemovedEvent{view}})
}

func (el *eventLog) log(jev interface{}) {
	buf, err := json.Marshal(jev)
	if err != nil {
		log.Errorf("Could not marshal event %+v: %s", jev, err)
		return
	}
	_, err = el.writer.Write([]byte(string(buf) + "\n"))
	if err != nil {
		log.Errorf("Error writing event: %s", err)
		return
	}
}

func New(cfg config.Config) service.ServiceCallback {
	var writer io.Writer
	var err error
	if len(cfg.Eventlog.Path) == 0 {
		writer = ioutil.Discard
	} else {
		writer, err = makeWriter(cfg)
		if err != nil {
			log.Panicf("Error opening event log for writing: %s", err)
		}
		log.Infof("Logging events to %s", cfg.Eventlog.Path)
	}

	return &eventLog{writer}
}
