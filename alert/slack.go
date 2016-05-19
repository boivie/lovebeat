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
	publicUrl string
	cfg       *config.ConfigSlack
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
			Text     string       `json:"text"`
			Fields   []SlackField `json:"fields"`
		}

		prevUpper := strings.ToUpper(ev.Previous)
		currentUpper := strings.ToUpper(ev.Current)
		view := ev.View

		var color string
		var title string
		link := m.publicUrl + "views/" + view.Name
		if ev.Current == model.StateError {
			color = "danger"
			title = fmt.Sprintf("<%s|%s> is in ERROR", link, view.Name)
		} else {
			color = "good"
			title = fmt.Sprintf("<%s|%s> has recovered", link, view.Name)
		}

		payload := struct {
			Username    string            `json:"username"`
			IconUrl     string            `json:"icon_url"`
			Channel     string            `json:"channel"`
			Attachments []SlackAttachment `json:"attachments"`
		}{
			Username: "Lovebeat",
			IconUrl:  "https://cdn.rawgit.com/boivie/lovebeat/a440f41aaf74be2e14b14a8f470456b71fc3e64e/docs/lovebeat-48.png",
			Channel:  cfg.SlackChannel,
			Attachments: []SlackAttachment{
				SlackAttachment{
					Fallback: fmt.Sprintf("lovebeat: %s changed from %s to %s", view.Name, prevUpper, currentUpper),
					Color:    color,
					Title:    title,
					Text:     fmt.Sprintf("Changed from %s to %s.", prevUpper, currentUpper),
					Fields: []SlackField{
						SlackField{Title: "View Name", Value: view.Name, Short: true},
						SlackField{Title: "Incident Number", Value: fmt.Sprintf("#%d", view.IncidentNbr), Short: true},
					},
				},
			},
		}

		if ev.Current == model.StateError {
			payload.Attachments[0].Fields = append(payload.Attachments[0].Fields,
				SlackField{Title: "Failed Service(s)", Value: strings.Join(ev.View.FailedServices, ", "), Short: false})
		}

		log.Debugf("Performing slack request at %v", m.cfg.WebhookUrl)
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
			log.Errorf("Failed to post slack alert: %v", err)
		}
	}
}

func NewSlackAlerter(cfg config.Config) AlerterBackend {
	return &slackAlerter{cfg.General.PublicUrl, &cfg.Slack}
}
