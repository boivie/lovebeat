package eventlog

import (
	"encoding/json"
	"github.com/boivie/lovebeat/eventbus"
	"github.com/boivie/lovebeat/service"
	"github.com/op/go-logging"
	"io"
	"reflect"
	"time"
)

var log = logging.MustGetLogger("lovebeat")

type EventLog struct {
	writer io.Writer
}

func New(writer io.Writer) *EventLog {
	return &EventLog{writer}
}

func (el *EventLog) Register(bus *eventbus.EventBus) {
	bus.RegisterHandler(
		el.eventHandler,
		service.ServiceStateChangedEvent{},
		service.ViewStateChangedEvent{})
}

func (el *EventLog) eventHandler(ev interface{}) {
	t := camelToSnakeCase(reflect.TypeOf(ev).Name())
	jev := map[string]interface{}{
		"ts":   time.Now().UTC(),
		"type": t,
		t:      &ev,
	}
	buf, err := json.Marshal(jev)
	if err != nil {
		log.Error("Could not marshal event %+v: %s", jev, err)
		return
	}
	_, err = el.writer.Write([]byte(string(buf) + "\n"))
	if err != nil {
		log.Error("Error writing event: %s", err)
		return
	}
}
