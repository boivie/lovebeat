package alert

import (
	"bytes"
	"github.com/boivie/lovebeat/config"
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
	cmds chan mail
}

const (
	TMPL_BODY = `The status for view '{{.Current.Name}}' has changed from '{{.Previous.State | ToUpper}}' to '{{.Current.State | ToUpper}}'

Services with failures (max 10):

{{range .Services}}  * {{.Name}} - {{.State | ToUpper}}
{{else}}  None. All are OK.{{end}}
`
	TMPL_SUBJECT = `[LOVEBEAT] {{.Current.Name}}-{{.Current.IncidentNbr}}`
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
		log.Error("error trying to parse mail template", err)
		return ""
	}
	var doc bytes.Buffer

	err = t.Execute(&doc, context)
	if err != nil {
		log.Error("Failed to render template", err)
		return ""
	}
	return doc.String()
}

func createMail(alert Alert) mail {
	var context = make(map[string]interface{})
	context["Previous"] = alert.Previous
	context["Current"] = alert.Current
	context["Services"] = alert.ServicesInError

	var body = renderTemplate(TMPL_BODY, context)
	var subject = renderTemplate(TMPL_SUBJECT, context)
	return mail{To: alert.Current.AlertMail,
		Subject: subject,
		Body:    body}
}

func (m mailAlerter) Notify(alert Alert) {
	if alert.Current.AlertMail != "" {
		m.cmds <- createMail(alert)
	}
}

func (m mailAlerter) Worker(q chan mail, cfg *config.ConfigMail) {
	for {
		select {
		case mail := <-q:
			log.Info("Sending from %s on host %s", cfg.From, cfg.Server)
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
				log.Error("Failed to send e-mail", err)
			}
		}
	}

}

func NewMailAlerter(cfg *config.ConfigMail) Alerter {
	log.Debug("Sending mail via %s, from %s", cfg.Server, cfg.From)
	var q = make(chan mail, 100)
	var ma = mailAlerter{cmds: q}
	go ma.Worker(q, cfg)
	return &ma
}
