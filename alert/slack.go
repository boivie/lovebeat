package alert

import (
	"fmt"
	"github.com/boivie/lovebeat/config"
	"github.com/boivie/lovebeat/model"
	"github.com/franela/goreq"
	"strings"
	"time"
)

type slackAlerter struct {
	cfg *config.ConfigSlack
}

func (m slackAlerter) Notify(cfg config.ConfigAlert, ev AlertInfo) {
	if cfg.SlackChannel != "" {
		type SlackField struct {
			Title string `json:"title"`
			Value string `json:"value"`
			Short bool   `json:"short"`
		}

		type SlackAttachment struct {
			Fallback string       `json:"fallback"`
			Color    string       `json:"color"`
			Title    string       `json:"title"`
			Fields   []SlackField `json:"fields"`
		}

		var color string
		if ev.Current == model.StateError {
			color = "danger"
		} else {
			color = "good"
		}

		prevUpper := strings.ToUpper(ev.Previous)
		currentUpper := strings.ToUpper(ev.Current)
		view := ev.View

		payload := struct {
			Username    string            `json:"username"`
			IconEmoji   string            `json:"icon_emoji"`
			Channel     string            `json:"channel"`
			Attachments []SlackAttachment `json:"attachments"`
		}{
			Username:  "Lovebeat",
			IconEmoji: ":loud_sound:",
			Channel:   cfg.SlackChannel,
			Attachments: []SlackAttachment{
				SlackAttachment{
					Fallback: fmt.Sprintf("lovebeat: %s changed from %s to %s", view.Name, prevUpper, currentUpper),
					Color:    color,
					Title:    fmt.Sprintf("\"%s\" has changed from %s to %s", view.Name, prevUpper, currentUpper),
					Fields: []SlackField{
						SlackField{Title: "View Name", Value: view.Name, Short: true},
						SlackField{Title: "Incident Number", Value: fmt.Sprintf("#%d", view.IncidentNbr), Short: true},
						SlackField{Title: "From State", Value: prevUpper, Short: true},
						SlackField{Title: "Failed Service(s)", Value: strings.Join(formatFailedServices(ev.FailedServices), ","), Short: true},
						SlackField{Title: "To State", Value: currentUpper, Short: true},
					},
				},
			},
		}

		log.Debug("Performing slack request at %v", m.cfg.WebhookUrl)
		goreq.SetConnectTimeout(5 * time.Second)
		req := goreq.Request{
			Method:      "POST",
			Uri:         m.cfg.WebhookUrl,
			Accept:      "application/json",
			ContentType: "application/json",
			UserAgent:   "Lovebeat",
			Timeout:     10 * time.Second,
			Body:        payload,
		}
		_, err := req.Do()

		if err != nil {
			log.Error("Failed to post slack alert: %v", err)
		}
	}
}

func formatFailedServices(services []model.Service) []string {
	var names = make([]string, 0)
	for _, service := range services {
		names = append(names, service.Name)
	}
	return names
}

func NewSlackAlerter(cfg config.Config) AlerterBackend {
	return &slackAlerter{&cfg.Slack}
}
