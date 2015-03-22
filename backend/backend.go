package backend

import (
	"github.com/boivie/lovebeat/model"
)

type Backend interface {
	LoadServices() []*model.Service
	LoadViews() []*model.View

	SaveService(service *model.Service)
	SaveView(view *model.View)

	DeleteService(name string)
	DeleteView(name string)
	Sync()
}
