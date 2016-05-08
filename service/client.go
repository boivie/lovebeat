package service

import (
	"github.com/boivie/lovebeat/model"
	"time"
)

const ServiceNamePattern = "[a-z0-9._-]+"

type getServicesCmd struct {
	View  string
	Reply chan []model.Service
}

type getServiceCmd struct {
	Name  string
	Reply chan *model.Service
}

type getViewsCmd struct {
	Reply chan []model.View
}

type getViewCmd struct {
	Name  string
	Reply chan *model.View
}

func tsnow() int64 {
	return int64(time.Now().UnixNano() / 1e6)
}

func (c *ServicesImpl) Update(update *Update) {
	c.updateChan <- update
}

func (c *ServicesImpl) GetServices(view string) []model.Service {
	myc := make(chan []model.Service)
	c.getServicesChan <- &getServicesCmd{View: view, Reply: myc}
	ret := <-myc
	return ret
}

func (c *ServicesImpl) GetService(name string) *model.Service {
	myc := make(chan *model.Service)
	c.getServiceChan <- &getServiceCmd{Name: name, Reply: myc}
	ret := <-myc
	return ret
}

func (c *ServicesImpl) GetViews() []model.View {
	myc := make(chan []model.View)
	c.getViewsChan <- &getViewsCmd{Reply: myc}
	ret := <-myc
	return ret
}

func (c *ServicesImpl) GetView(name string) *model.View {
	myc := make(chan *model.View)
	c.getViewChan <- &getViewCmd{Name: name, Reply: myc}
	ret := <-myc
	return ret
}
