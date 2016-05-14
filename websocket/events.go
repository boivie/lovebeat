package websocket

import (
	"encoding/json"
	"github.com/boivie/lovebeat/model"
)

type webSocketEvent struct {
	Type    string         `json:"type"`
	Service *model.Service `json:"service,omitempty"`
	View    *model.View    `json:"view,omitempty"`
}

func serviceAdded(ev model.ServiceAddedEvent) {
	var encoded, _ = json.Marshal(webSocketEvent{Type: "ADD_SERVICE", Service: &ev.Service})
	h.broadcast <- encoded
}

func serviceStateChanged(ev model.ServiceStateChangedEvent) {
	var encoded, _ = json.Marshal(webSocketEvent{Type: "UPDATE_SERVICE", Service: &ev.Service})
	h.broadcast <- encoded
}

func serviceRemoved(ev model.ServiceRemovedEvent) {
	var encoded, _ = json.Marshal(webSocketEvent{Type: "REMOVE_SERVICE", Service: &ev.Service})
	h.broadcast <- encoded
}

func viewAdded(ev model.ViewAddedEvent) {
	var encoded, _ = json.Marshal(webSocketEvent{Type: "ADD_VIEW", View: &ev.View})
	h.broadcast <- encoded
}

func viewStateChanged(ev model.ViewStateChangedEvent) {
	var encoded, _ = json.Marshal(webSocketEvent{Type: "UPDATE_VIEW", View: &ev.View})
	h.broadcast <- encoded
}
