package alert

import (
	"bytes"
	"github.com/boivie/lovebeat/config"
	"gopkg.in/mailgun/mailgun-go.v1"
	"net/smtp"
	"strings"
	"text/template"
)

type mail struct {
	To      string
	Subject string
	Body    string
}

type mailAlerter struct {
	smtp    *config.ConfigSmtp
	mailgun *config.ConfigMailgun
}

const (
	TMPL_BODY = `The status for alarm '{{.Alarm.Name}}' has changed from '{{.Previous | ToUpper}}' to '{{.Current | ToUpper}}'
`
	TMPL_SUBJECT = `[LOVEBEAT] {{.Alarm.Name}}-{{.Alarm.IncidentNbr}}`
	TMPL_EMAIL   = `From: {{.From}}
To: {{.To}}
Subject: {{.Subject}}
MIME-version: 1.0
Content-Type: text/html; charset="UTF-8"

{{.Message}}`
)

func renderTemplate(tmpl string, context map[string]interface{}) string {
	funcMap := template.FuncMap{
		"ToUpper": strings.ToUpper,
	}
	t, err := template.New("template").Funcs(funcMap).Parse(tmpl)
	if err != nil {
		log.Errorf("error trying to parse mail template: %s", err)
		return ""
	}
	var doc bytes.Buffer

	err = t.Execute(&doc, context)
	if err != nil {
		log.Errorf("Failed to render template: %s", err)
		return ""
	}
	return doc.String()
}

func createMail(address string, ev AlertInfo) mail {
	var context = make(map[string]interface{})
	context["Alarm"] = ev.Alarm
	context["Previous"] = ev.Previous
	context["Current"] = ev.Current

	var body = renderTemplate(TMPL_BODY, context)
	var subject = renderTemplate(TMPL_SUBJECT, context)
	return mail{
		To:      address,
		Subject: subject,
		Body:    body,
	}
}

func (m mailAlerter) Notify(cfg config.ConfigAlert, ev AlertInfo) {
	if cfg.Mail != "" {
		mail := createMail(cfg.Mail, ev)

		if m.mailgun != nil {
			sendMailgun(m.mailgun, mail)
		} else {
			sendSmtp(m.smtp, mail)
		}
	}
}

func sendSmtp(cfg *config.ConfigSmtp, mail mail) {
	log.Infof("Sending from %s via SMTP server %s", cfg.From, cfg.Server)
	var context = make(map[string]interface{})
	context["From"] = cfg.From
	context["To"] = mail.To
	context["Subject"] = mail.Subject
	context["Message"] = mail.Body

	contents := renderTemplate(TMPL_EMAIL, context)
	var to = strings.Split(mail.To, ",")
	var err = smtp.SendMail(cfg.Server, nil, cfg.From, to,
		[]byte(contents))
	if err != nil {
		log.Errorf("Failed to send e-mail: %s", err)
	}
}

func sendMailgun(cfg *config.ConfigMailgun, mail mail) {
	log.Infof("Sending from %s via mailgun domain %s", cfg.From, cfg.Domain)

	mg := mailgun.NewMailgun(cfg.Domain, cfg.ApiKey, "")
	message := mailgun.NewMessage(cfg.From, mail.Subject, mail.Body, mail.To)

	_, id, err := mg.Send(message)
	if err != nil {
		log.Errorf("Failed to send e-mail: %s", err)
	} else {
		log.Infof("Sent mail as ID %s", id)
	}
}

func NewMailAlerter(cfg config.Config) AlerterBackend {
	if cfg.Mailgun.Domain != "" {
		log.Infof("Sending mail via Mailgun from domain %s, from %s", cfg.Mailgun.Domain, cfg.Mailgun.From)
		return &mailAlerter{mailgun: &cfg.Mailgun}
	}

	log.Infof("Sending mail via SMTP at %s, from %s", cfg.Mail.Server, cfg.Mail.From)
	return &mailAlerter{smtp: &cfg.Mail}
}
