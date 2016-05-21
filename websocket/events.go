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
	View    *model.View         `json:"view,omitempty"`
}

type wsObj struct {
}

func (s *wsObj) OnUpdate(ts int64, update model.Update) {}

func (s *wsObj) OnServiceAdded(ts int64, service model.Service) {
	var encoded, _ = json.Marshal(webSocketEvent{Type: "ADD_SERVICE", Service: api.ToHttpService(service)})
	h.broadcast <- encoded
}
func (s *wsObj) OnServiceUpdated(ts int64, oldService, newService model.Service) {
	var encoded, _ = json.Marshal(webSocketEvent{Type: "UPDATE_SERVICE", Service: api.ToHttpService(newService)})
	h.broadcast <- encoded
}
func (s *wsObj) OnServiceRemoved(ts int64, service model.Service) {
	var encoded, _ = json.Marshal(webSocketEvent{Type: "REMOVE_SERVICE", Service: api.ToHttpService(service)})
	h.broadcast <- encoded
}

func (s *wsObj) OnViewAdded(ts int64, view model.View, config config.ConfigView) {
	var encoded, _ = json.Marshal(webSocketEvent{Type: "ADD_VIEW", View: &view})
	h.broadcast <- encoded
}
func (s *wsObj) OnViewUpdated(ts int64, oldView, newView model.View, config config.ConfigView) {
	var encoded, _ = json.Marshal(webSocketEvent{Type: "UPDATE_VIEW", View: &newView})
	h.broadcast <- encoded
}
func (s *wsObj) OnViewRemoved(ts int64, view model.View, config config.ConfigView) {
	var encoded, _ = json.Marshal(webSocketEvent{Type: "REMOVE_VIEW", View: &view})
	h.broadcast <- encoded
}
