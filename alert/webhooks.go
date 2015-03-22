package alert

import (
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

func (m webhooksAlerter) Notify(url string, alert Alert) {
	js := webhookData{Name: alert.Current.Name,
		FromState:   strings.ToUpper(alert.Previous.State),
		ToState:     strings.ToUpper(alert.Current.State),
		IncidentNbr: alert.Current.IncidentNbr}
	m.cmds <- webhook{Url: url, Data: js}
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

/*
func NewWebhooksAlerter() Alerter {
	goreq.SetConnectTimeout(5 * time.Second)
	var q = make(chan webhook, 100)
	var w = webhooksAlerter{cmds: q}
	go w.Worker(q)
	return &w
}
*/
