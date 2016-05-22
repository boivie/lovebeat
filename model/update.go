package model

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

type DeleteAlarm struct {
}

type Update struct {
	Ts int64 `json:"ts"`

	Tick *Tick `json:"tick,omitempty"`

	Service       string         `json:"service,omitempty"`
	Beat          *Beat          `json:"beat,omitempty"`
	SetTimeout    *SetTimeout    `json:"set_timeout,omitempty"`
	MuteService   *MuteService   `json:"mute_service,omitempty"`
	DeleteService *DeleteService `json:"delete_service,omitempty"`

	Alarm       string       `json:"alarm,omitempty"`
	DeleteAlarm *DeleteAlarm `json:"delete_alarm,omitempty"`
}
