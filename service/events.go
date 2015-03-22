package service

import (
	"github.com/boivie/lovebeat/model"
)

type ViewStateChangedEvent struct {
	View     model.View
	Previous string
	Current  string
}

type ServiceStateChangedEvent struct {
	Service  model.Service
	Previous string
	Current  string
}
