package alert

import (
	"github.com/boivie/lovebeat/config"
	"github.com/franela/goreq"
	"strings"
	"time"
)

type webhooksAlerter struct{}

func (m webhooksAlerter) Notify(cfg config.ConfigAlert, ev AlertInfo) {
	if cfg.Webhook != "" {
		log.Infof("Sending webhook alert to %s", cfg.Webhook)

		req := goreq.Request{
			Uri:         cfg.Webhook,
			Accept:      "application/json",
			ContentType: "application/json",
			UserAgent:   "Lovebeat",
			Timeout:     10 * time.Second,
			Body: struct {
				Name        string `json:"name"`
				FromState   string `json:"from_state"`
				ToState     string `json:"to_state"`
				IncidentNbr int    `json:"incident_number"`
			}{
				Name:        ev.Alarm.Name,
				FromState:   strings.ToUpper(ev.Previous),
				ToState:     strings.ToUpper(ev.Current),
				IncidentNbr: ev.Alarm.IncidentNbr,
			},
		}

		req.AddHeader("X-Lovebeat", "1")

		_, err := req.Do()
		if err != nil {
			log.Errorf("Failed to post webhook: %s", err)
		}
	}
}

func NewWebhooksAlerter(cfg config.Config) AlerterBackend {
	return &webhooksAlerter{}
}
