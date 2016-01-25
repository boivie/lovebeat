package alert

import (
	"github.com/boivie/lovebeat/config"
	"github.com/boivie/lovebeat/service"
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

func (m webhooksAlerter) Notify(cfg config.ConfigAlert, ev service.ViewStateChangedEvent) {
	if cfg.Webhook != "" {
		js := webhookData{
			Name:        ev.View.Name,
			FromState:   strings.ToUpper(ev.Previous),
			ToState:     strings.ToUpper(ev.Current),
			IncidentNbr: ev.View.IncidentNbr}
		m.cmds <- webhook{Url: cfg.Webhook, Data: js}
	}
}

func (m webhooksAlerter) Worker(q chan webhook) {
	for {
		webhook := <-q
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

func NewWebhooksAlerter(cfg config.Config) Alerter {
	goreq.SetConnectTimeout(5 * time.Second)
	var q = make(chan webhook, 100)
	var w = webhooksAlerter{cmds: q}
	go w.Worker(q)
	return &w
}
