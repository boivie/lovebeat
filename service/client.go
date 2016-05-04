package service

import (
	"github.com/boivie/lovebeat/model"
)

const ServiceNamePattern = "[a-z0-9._-]+"

type ServiceIf interface {
	DeleteService(name string)

	// Updates the service. The timeouts are in milliseconds.
	UpdateService(name string, registerBeat bool, changeTimeout bool, timeout int64)
	MuteService(name string, muted bool)

	GetServices(view string) []model.Service
	GetService(name string) *model.Service
	GetViews() []model.View
	GetView(name string) *model.View
}

type upsertServiceCmd struct {
	RegisterBeat bool
	Service      string
	HasTimeout   bool
	Timeout      int64
}

type muteServiceCmd struct {
	Service string
	Muted   bool
}

type upsertViewCmd struct {
	View      string
	Regexp    string
	AlertMail string
	Webhooks  string
}

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

type client struct {
	svcs *Services
}

func (c *client) DeleteService(name string) {
	c.svcs.deleteServiceCmdChan <- name
}

func (c *client) UpdateService(name string, registerBeat bool, changeTimeout bool, timeout int64) {
	c.svcs.upsertServiceCmdChan <- &upsertServiceCmd{
		Service:      name,
		RegisterBeat: registerBeat,
		HasTimeout:   changeTimeout,
		Timeout:      timeout,
	}
}

func (c *client) MuteService(name string, muted bool) {
	c.svcs.muteServiceCmdChan <- &muteServiceCmd{name, muted}
}

func (c *client) GetServices(view string) []model.Service {
	myc := make(chan []model.Service)
	c.svcs.getServicesChan <- &getServicesCmd{View: view, Reply: myc}
	ret := <-myc
	return ret
}

func (c *client) GetService(name string) *model.Service {
	myc := make(chan *model.Service)
	c.svcs.getServiceChan <- &getServiceCmd{Name: name, Reply: myc}
	ret := <-myc
	return ret
}

func (c *client) GetViews() []model.View {
	myc := make(chan []model.View)
	c.svcs.getViewsChan <- &getViewsCmd{Reply: myc}
	ret := <-myc
	return ret
}

func (c *client) GetView(name string) *model.View {
	myc := make(chan *model.View)
	c.svcs.getViewChan <- &getViewCmd{Name: name, Reply: myc}
	ret := <-myc
	return ret
}

func (svcs *Services) GetClient() ServiceIf {
	return &client{svcs: svcs}
}
