package alert

import (
	"github.com/boivie/lovebeat/config"
	"github.com/boivie/lovebeat/service"
	"github.com/franela/goreq"
	"time"
	"fmt"
	"github.com/boivie/lovebeat/model"
	"strings"
)

type slackAlert struct {
	Channel string
	Data    service.ViewStateChangedEvent
}

type slackAlerter struct {
	cmds chan slackAlert
}

func (m slackAlerter) Notify(cfg config.ConfigAlert, ev service.ViewStateChangedEvent) {
	if cfg.SlackChannel != "" {
		m.cmds <- slackAlert{Channel: cfg.SlackChannel, Data: ev}
	}
}

func (m slackAlerter) Worker(q <- chan slackAlert, cfg *config.ConfigSlack) {
	for slackAlert := range q {
		type SlackField struct {
			Title string `json:"title"`
			Value string `json:"value"`
			Short bool `json:"short"`
		}

		type SlackAttachment struct {
			Fallback string `json:"fallback"`
			Color    string `json:"color"`
			Title    string `json:"title"`
			Fields   []SlackField `json:"fields"`
		}

		var color string
		if slackAlert.Data.Current == model.StateError {
			color = "danger"
		} else {
			color = "good"
		}

		prevUpper := strings.ToUpper(slackAlert.Data.Previous)
		currentUpper := strings.ToUpper(slackAlert.Data.Current)
		view := slackAlert.Data.View

		payload := struct {
			Username    string `json:"username"`
			IconEmoji   string `json:"icon_emoji"`
			Channel     string `json:"channel"`
			Attachments []SlackAttachment `json:"attachments"`
		}{
			Username: "Lovebeat",
			IconEmoji: ":loud_sound:",
			Channel: slackAlert.Channel,
			Attachments: []SlackAttachment{
				SlackAttachment{
					Fallback: fmt.Sprintf("lovebeat: %s changed from %s to %s", view.Name, prevUpper, currentUpper),
					Color: color,
					Title: fmt.Sprintf("\"%s\" has changed from %s to %s", view.Name, prevUpper, currentUpper),
					Fields: []SlackField{
						SlackField{Title: "View Name", Value: view.Name, Short: true },
						SlackField{Title: "Incident Number", Value: fmt.Sprintf("#%d", view.IncidentNbr), Short: true },
						SlackField{Title: "From State", Value: prevUpper, Short: true },
						SlackField{Title: "Failed Service(s)", Value: strings.Join(formatFailedServices(slackAlert.Data)[:], ","), Short: true },
						SlackField{Title: "To State", Value: currentUpper, Short: true },
					},
				},
			},
		}

		log.Debug("Performing slack request at %v", cfg.WebhookUrl)
		req := goreq.Request{
			Method:      "POST",
			Uri:         cfg.WebhookUrl,
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

func formatFailedServices(event service.ViewStateChangedEvent) []string {
	var services = make([]string, 0)
	for _, service := range event.FailedServices {
		services = append(services, fmt.Sprintf("%s (%s)", service.Name, service.State))
	}
	return services
}

func NewSlackAlerter(cfg config.Config) Alerter {
	goreq.SetConnectTimeout(5 * time.Second)

	var q = make(chan slackAlert, 100)
	var w = slackAlerter{cmds: q}
	go w.Worker(q, &cfg.Slack)
	return &w
}
