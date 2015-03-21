package service

import (
	"github.com/boivie/lovebeat/backend"
)

// Special values for error and warning timeouts
const TIMEOUT_CLEAR int64 = -1
const TIMEOUT_AUTO int64 = -2

type ServiceIf interface {
	DeleteService(name string)

	UpdateService(name string, registerBeat bool, warningTimeout int64, errorTimeout int64)

	CreateOrUpdateView(name string, regexp string, alertMail string, webhooks string)
	DeleteView(name string)
	GetServices(view string) []backend.StoredService
	GetService(name string) *backend.StoredService
	GetViews() []backend.StoredView
	GetView(name string) *backend.StoredView
}

type upsertServiceCmd struct {
	RegisterBeat   bool
	Service        string
	WarningTimeout int64
	ErrorTimeout   int64
}

type upsertViewCmd struct {
	View      string
	Regexp    string
	AlertMail string
	Webhooks  string
}

type getServicesCmd struct {
	View  string
	Reply chan []backend.StoredService
}

type getServiceCmd struct {
	Name  string
	Reply chan *backend.StoredService
}

type getViewsCmd struct {
	Reply chan []backend.StoredView
}

type getViewCmd struct {
	Name  string
	Reply chan *backend.StoredView
}

type client struct {
	svcs *Services
}

func (c *client) DeleteService(name string) {
	c.svcs.deleteServiceCmdChan <- name
}

func (c *client) DeleteView(name string) {
	c.svcs.deleteViewCmdChan <- name
}

func (c *client) UpdateService(name string, registerBeat bool, warningTimeout int64, errorTimeout int64) {
	c.svcs.upsertServiceCmdChan <- &upsertServiceCmd{
		Service:        name,
		RegisterBeat:   registerBeat,
		WarningTimeout: warningTimeout,
		ErrorTimeout:   errorTimeout,
	}
}

func (c *client) CreateOrUpdateView(name string, regexp string, alertMail string, webhooks string) {
	c.svcs.upsertViewCmdChan <- &upsertViewCmd{
		View:      name,
		Regexp:    regexp,
		AlertMail: alertMail,
		Webhooks:  webhooks,
	}
}

func (c *client) GetServices(view string) []backend.StoredService {
	myc := make(chan []backend.StoredService)
	c.svcs.getServicesChan <- &getServicesCmd{View: view, Reply: myc}
	ret := <-myc
	return ret
}

func (c *client) GetService(name string) *backend.StoredService {
	myc := make(chan *backend.StoredService)
	c.svcs.getServiceChan <- &getServiceCmd{Name: name, Reply: myc}
	ret := <-myc
	return ret
}

func (c *client) GetViews() []backend.StoredView {
	myc := make(chan []backend.StoredView)
	c.svcs.getViewsChan <- &getViewsCmd{Reply: myc}
	ret := <-myc
	return ret
}

func (c *client) GetView(name string) *backend.StoredView {
	myc := make(chan *backend.StoredView)
	c.svcs.getViewChan <- &getViewCmd{Name: name, Reply: myc}
	ret := <-myc
	return ret
}

func (svcs *Services) GetClient() ServiceIf {
	return &client{svcs: svcs}
}
