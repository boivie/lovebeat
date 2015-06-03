package eventbus

import (
	"fmt"
	"reflect"
	"sync"
)

type EventBus struct {
	handlers map[reflect.Type][]reflect.Value
	lock     sync.RWMutex
}

func New() *EventBus {
	return &EventBus{
		make(map[reflect.Type][]reflect.Value),
		sync.RWMutex{},
	}
}

func (bus *EventBus) addHandler(t reflect.Type, fn reflect.Value) {
	bus.lock.Lock()
	defer bus.lock.Unlock()
	handlers, ok := bus.handlers[t]
	if !ok {
		handlers = make([]reflect.Value, 0)
	}
	bus.handlers[t] = append(handlers, fn)
}

func (bus *EventBus) RegisterHandler(fn interface{}, forTypes ...interface{}) {
	v := reflect.ValueOf(fn)
	def := v.Type()

	// the message handler must have a single parameter
	if def.NumIn() != 1 {
		panic("Handler must have a single argument")
	}
	// find out the handler argument type
	argument := def.In(0)

	// check wether we can convert the types into the argument
	for _, typ := range forTypes {
		t := reflect.TypeOf(typ)
		if !t.ConvertibleTo(argument) {
			panic(fmt.Sprintf("Handler argument %v is not compatible with type %v", argument, t))
		}
		bus.addHandler(t, v)
	}

	// if we aren't specific, we just handle the specified message
	if len(forTypes) == 0 {
		bus.addHandler(argument, v)
	}
}

func (bus *EventBus) Publish(ev interface{}) error {
	bus.lock.RLock()
	defer bus.lock.RUnlock()

	t := reflect.TypeOf(ev)

	handlers, ok := bus.handlers[t]
	if !ok {
		return nil
	}

	args := [...]reflect.Value{reflect.ValueOf(ev)}
	for _, fn := range handlers {
		fn.Call(args[:])
	}
	return nil
}
