package alert

import (
	"bytes"
	"github.com/boivie/lovebeat-go/backend"
	"github.com/boivie/lovebeat-go/config"
	"github.com/op/go-logging"
	"net/smtp"
	"strconv"
	"strings"
	"text/template"
)

var (
	log = logging.MustGetLogger("lovebeat")
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
	TMPL_BODY    = `The status for view '{{.Name}}' has changed from '{{.PrevState}}' to '{{.CurrentState}}'`
	TMPL_SUBJECT = `[LOVEBEAT][{{.Name}}-{{.IncidentNbr}}]`
)

func renderTemplate(tmpl string, context map[string]string) string {
	t := template.New("template")
	var err error
	t, err = t.Parse(tmpl)
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

func (m mailAlerter) Notify(previous backend.StoredView,
	current backend.StoredView,
	servicesInError []backend.StoredService) {
	if current.AlertMail != "" {
		var context = make(map[string]string)
		context["Name"] = current.Name
		context["PrevState"] = strings.ToUpper(previous.State)
		context["CurrentState"] = strings.ToUpper(current.State)
		context["IncidentNbr"] = strconv.Itoa(current.IncidentNbr)

		var body = renderTemplate(TMPL_BODY, context)
		var subject = renderTemplate(TMPL_SUBJECT, context)
		m.cmds <- mail{To: current.AlertMail,
			Subject: subject,
			Body:    body}
	}
}

func (m mailAlerter) Worker(q chan mail, cfg *config.ConfigMail) {
	for {
		select {
		case mail := <-q:
			log.Info("Sending from %s on host %s", cfg.From, cfg.Server)
			parameters := struct {
				From    string
				To      string
				Subject string
				Message string
			}{
				cfg.From,
				mail.To,
				mail.Subject,
				mail.Body,
			}

			buffer := new(bytes.Buffer)

			template := template.Must(template.New("emailTemplate").Parse(emailScript()))
			template.Execute(buffer, &parameters)

			var to = strings.Split(mail.To, ",")
			var err = smtp.SendMail(cfg.Server, nil, cfg.From, to, buffer.Bytes())
			if err != nil {
				log.Error("Failed to send e-mail", err)
			}
		}
	}

}

func emailScript() (script string) {
	return `From: {{.From}}
To: {{.To}}
Subject: {{.Subject}}
MIME-version: 1.0
Content-Type: text/html; charset="UTF-8"

{{.Message}}`
}

func NewMailAlerter(cfg *config.ConfigMail) Alerter {
	var q = make(chan mail, 100)
	var ma = mailAlerter{cmds: q}
	go ma.Worker(q, cfg)
	return &ma
}
