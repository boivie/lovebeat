package service

import (
	"github.com/boivie/lovebeat/config"
	"github.com/boivie/lovebeat/model"
)

type Services interface {
	Subscribe(cb ServiceCallback)
	Update(update *model.Update)

	GetServices(view string) []model.Service
	GetService(name string) *model.Service
	GetViews() []model.View
	GetView(name string) *model.View
}

type ServiceCallback interface {
	OnUpdate(ts int64, update model.Update)

	OnServiceAdded(ts int64, service model.Service)
	OnServiceUpdated(ts int64, oldService, newService model.Service)
	OnServiceRemoved(ts int64, service model.Service)

	OnViewAdded(ts int64, view model.View, config config.ConfigView)
	OnViewUpdated(ts int64, oldView, newView model.View, config config.ConfigView)
	OnViewRemoved(ts int64, view model.View, config config.ConfigView)
}
