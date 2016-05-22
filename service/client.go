package service

import (
	"github.com/boivie/lovebeat/model"
	"time"
)

const ServiceNamePattern = "[a-z0-9._-]+"

type getServicesInAlarmCmd struct {
	Alarm string
	Reply chan []model.Service
}

type getServiceCmd struct {
	Name  string
	Reply chan *model.Service
}

type getAlarmsCmd struct {
	Reply chan []model.Alarm
}

type getAlarmCmd struct {
	Name  string
	Reply chan *model.Alarm
}

func tsnow() int64 {
	return int64(time.Now().UnixNano() / 1e6)
}

func (c *ServicesImpl) Subscribe(cb ServiceCallback) {
	c.subscribeChan <- cb
}

func (c *ServicesImpl) Update(update *model.Update) {
	c.updateChan <- update
}

func (c *ServicesImpl) GetServicesInAlarm(alarm string) []model.Service {
	myc := make(chan []model.Service)
	c.getServicesChan <- &getServicesInAlarmCmd{Alarm: alarm, Reply: myc}
	ret := <-myc
	return ret
}

func (c *ServicesImpl) GetServices() []model.Service {
	myc := make(chan []model.Service)
	c.getServicesChan <- &getServicesInAlarmCmd{Alarm: "", Reply: myc}
	ret := <-myc
	return ret
}

func (c *ServicesImpl) GetService(name string) *model.Service {
	myc := make(chan *model.Service)
	c.getServiceChan <- &getServiceCmd{Name: name, Reply: myc}
	ret := <-myc
	return ret
}

func (c *ServicesImpl) GetAlarms() []model.Alarm {
	myc := make(chan []model.Alarm)
	c.getAlarmsChan <- &getAlarmsCmd{Reply: myc}
	ret := <-myc
	return ret
}

func (c *ServicesImpl) GetAlarm(name string) *model.Alarm {
	myc := make(chan *model.Alarm)
	c.getAlarmChan <- &getAlarmCmd{Name: name, Reply: myc}
	ret := <-myc
	return ret
}
