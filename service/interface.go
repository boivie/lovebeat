package service

import "github.com/boivie/lovebeat/model"

type Beat struct {
}

type SetTimeout struct {
	Timeout int64 `json:"timeout"`
}

type MuteService struct {
	Muted bool `json:"muted"`
}

type DeleteService struct {
}

type Tick struct {
}

type Update struct {
	Ts            int64          `json:"ts"`
	Service       string         `json:"service,omitempty"`
	Beat          *Beat          `json:"beat,omitempty"`
	SetTimeout    *SetTimeout    `json:"set_timeout,omitempty"`
	MuteService   *MuteService   `json:"mute_service,omitempty"`
	DeleteService *DeleteService `json:"delete_service,omitempty"`
	Tick          *Tick          `json:"tick,omitempty"`
}

type Services interface {
	Update(update *Update)

	GetServices(view string) []model.Service
	GetService(name string) *model.Service
	GetViews() []model.View
	GetView(name string) *model.View
}
