package service

import (
	"github.com/boivie/lovebeat/model"
)

type ViewStateChangedEvent struct {
	View     model.View `json:"view"`
	Previous string     `json:"previous"`
	Current  string     `json:"current"`
}

type ServiceStateChangedEvent struct {
	Service  model.Service `json:"service"`
	Previous string        `json:"previous"`
	Current  string        `json:"current"`
}

// When adding a new expression struct type here, don't forget
// to add it to the test cases so the member names are checked
// for conformity.
