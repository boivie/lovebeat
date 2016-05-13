package websocket

import (
	"encoding/json"
	"github.com/boivie/lovebeat/model"
)

func serviceStateChanged(ev model.ServiceStateChangedEvent) {
	var encoded, _ = json.Marshal(ev)
	h.broadcast <- encoded
}

func viewStateChanged(ev model.ViewStateChangedEvent) {
	var encoded, _ = json.Marshal(ev)
	h.broadcast <- encoded
}
