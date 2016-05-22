package websocket

import (
	"encoding/json"
	"github.com/boivie/lovebeat/api"
	"github.com/boivie/lovebeat/config"
	"github.com/boivie/lovebeat/model"
)

type webSocketEvent struct {
	Type    string              `json:"type"`
	Service *api.HttpApiService `json:"service,omitempty"`
	Alarm   *model.Alarm        `json:"alarm,omitempty"`
}

type wsObj struct {
}

func (s *wsObj) OnUpdate(ts int64, update model.Update) {}

func (s *wsObj) OnServiceAdded(ts int64, service model.Service) {
	srv := api.ToHttpService(service)
	var encoded, _ = json.Marshal(webSocketEvent{Type: "ADD_SERVICE", Service: &srv})
	h.broadcast <- encoded
}
func (s *wsObj) OnServiceUpdated(ts int64, oldService, newService model.Service) {
	srv := api.ToHttpService(newService)
	var encoded, _ = json.Marshal(webSocketEvent{Type: "UPDATE_SERVICE", Service: &srv})
	h.broadcast <- encoded
}
func (s *wsObj) OnServiceRemoved(ts int64, service model.Service) {
	srv := api.ToHttpService(service)
	var encoded, _ = json.Marshal(webSocketEvent{Type: "REMOVE_SERVICE", Service: &srv})
	h.broadcast <- encoded
}

func (s *wsObj) OnAlarmAdded(ts int64, alarm model.Alarm, config config.ConfigAlarm) {
	var encoded, _ = json.Marshal(webSocketEvent{Type: "ADD_ALARM", Alarm: &alarm})
	h.broadcast <- encoded
}
func (s *wsObj) OnAlarmUpdated(ts int64, oldAlarm, newAlarm model.Alarm, config config.ConfigAlarm) {
	var encoded, _ = json.Marshal(webSocketEvent{Type: "UPDATE_ALARM", Alarm: &newAlarm})
	h.broadcast <- encoded
}
func (s *wsObj) OnAlarmRemoved(ts int64, alarm model.Alarm, config config.ConfigAlarm) {
	var encoded, _ = json.Marshal(webSocketEvent{Type: "REMOVE_ALARM", Alarm: &alarm})
	h.broadcast <- encoded
}
