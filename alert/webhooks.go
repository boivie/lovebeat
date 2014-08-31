package alert

import (
	"github.com/boivie/lovebeat-go/backend"
	"github.com/franela/goreq"
	"strings"
	"time"
)

type webhook struct {
	Url  string
	Data webhookData
}

type webhooksAlerter struct {
	cmds chan webhook
}

type webhookData struct {
	Name        string `json:"name"`
	FromState   string `json:"from_state"`
	ToState     string `json:"to_state"`
	IncidentNbr int    `json:"incident_number"`
}

func (m webhooksAlerter) Notify(previous backend.StoredView,
	current backend.StoredView,
	servicesInError []backend.StoredService) {
	if current.Webhooks != "" {
		js := webhookData{Name: current.Name,
			FromState:   strings.ToUpper(previous.State),
			ToState:     strings.ToUpper(current.State),
			IncidentNbr: current.IncidentNbr}
		m.cmds <- webhook{Url: current.Webhooks, Data: js}
	}
}

func (m webhooksAlerter) Worker(q chan webhook) {
	for {
		select {
		case webhook := <-q:
			log.Info("Sending webhook alert to %s", webhook.Url)

			req := goreq.Request{
				Uri:         webhook.Url,
				Accept:      "application/json",
				ContentType: "application/json",
				UserAgent:   "Lovebeat",
				Timeout:     10 * time.Second,
				Body:        webhook.Data,
			}

			req.AddHeader("X-Lovebeat", "1")

			_, err := req.Do()
			if err != nil {
				log.Error("Failed to post webhook: %s", err)
			}
		}
	}

}

func NewWebhooksAlerter() Alerter {
	goreq.SetConnectTimeout(5 * time.Second)
	var q = make(chan webhook, 100)
	var w = webhooksAlerter{cmds: q}
	go w.Worker(q)
	return &w
}
