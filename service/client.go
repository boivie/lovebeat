package service

import (
	"github.com/boivie/lovebeat-go/backend"
)

type ServiceIf interface {
	Beat(name string)
	DeleteService(name string)

	SetWarningTimeout(name string, timeout int)
	SetErrorTimeout(name string, timeout int)

	CreateOrUpdateView(name string, regexp string)
	DeleteView(name string)
	GetServices(view string) []backend.StoredService
}

const (
	ACTION_SET_WARN = "set-warn"
	ACTION_SET_ERR  = "set-err"
	ACTION_BEAT     = "beat"
	ACTION_DELETE   = "delete"
)

const (
	ACTION_REFRESH_VIEW = "refresh-view"
	ACTION_UPSERT_VIEW  = "upsert-view"
	ACTION_DELETE_VIEW  = "delete"
)

type serviceCmd struct {
	Action  string
	Service string
	Value   int
}

type viewCmd struct {
	Action string
	View   string
	Regexp string
}

type getServicesCmd struct {
	View  string
	Reply chan []backend.StoredService
}

type client struct {
	svcs *Services
}

func (c *client) Beat(name string) {
	c.svcs.serviceCmdChan <- &serviceCmd{
		Action:  ACTION_BEAT,
		Service: name,
		Value:   1,
	}
}

func (c *client) DeleteService(name string) {
	c.svcs.serviceCmdChan <- &serviceCmd{
		Action:  ACTION_DELETE,
		Service: name,
	}
}

func (c *client) DeleteView(name string) {
	c.svcs.viewCmdChan <- &viewCmd{
		Action: ACTION_DELETE_VIEW,
		View:   name,
	}
}

func (c *client) SetWarningTimeout(name string, timeout int) {
	c.svcs.serviceCmdChan <- &serviceCmd{
		Action:  ACTION_SET_WARN,
		Service: name,
		Value:   timeout,
	}
}
func (c *client) SetErrorTimeout(name string, timeout int) {
	c.svcs.serviceCmdChan <- &serviceCmd{
		Action:  ACTION_SET_ERR,
		Service: name,
		Value:   timeout,
	}
}

func (c *client) CreateOrUpdateView(name string, regexp string) {
	c.svcs.viewCmdChan <- &viewCmd{
		Action: ACTION_UPSERT_VIEW,
		View:   name,
		Regexp: regexp,
	}
}

func (c *client) GetServices(view string) []backend.StoredService {
	myc := make(chan []backend.StoredService)
	c.svcs.getServicesChan <- &getServicesCmd{View: view, Reply: myc}
	ret := <-myc
	return ret
}

func (svcs *Services) GetClient() ServiceIf {
	return &client{svcs: svcs}
}
