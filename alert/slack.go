package alert

import (
	"bytes"
	"net/http"
	"io/ioutil"
	"github.com/boivie/lovebeat/config"
	"github.com/boivie/lovebeat/service"
	"github.com/franela/goreq"
	"text/template"
	"time"
	"net/url"
)

type slackhook struct {
	Uri  string
	Data service.ViewStateChangedEvent
}

type slackhookAlerter struct {
	cmds chan slackhook
}

func (m slackhookAlerter) Notify(cfg config.ConfigAlert, ev service.ViewStateChangedEvent) {

	if cfg.Slackhook != "" {

		m.cmds <- slackhook{Uri: cfg.Slackhook, Data: ev }
	}

}

func (m slackhookAlerter) Worker(q chan slackhook, cfg *config.ConfigSlackhook) {
	for {

		select {
		case slackhook := <-q:

			var err error

			var context = make(map[string]interface{})
			context["View"] = slackhook.Data.View
			context["Previous"] = slackhook.Data.Previous
			context["Current"] = slackhook.Data.Current

			tmpl := cfg.Template

			t, err := template.New("template").Parse(tmpl)
			if err != nil {
				log.Error("error trying to parse slackhook template:%s:err:%v:", tmpl, err)
				return
			}
			var doc bytes.Buffer

			err = t.Execute(&doc, context)
			if err != nil {
				log.Error("Failed to render template", err)
				return
			}

			req := goreq.Request{
			    Method: "POST",
				Uri:         cfg.Uri,
				Accept:      "*/*",
				ContentType: "application/x-www-form-urlencoded",
				UserAgent:   "Lovebeat",
				Timeout:     10 * time.Second,
				Body:        "payload=" + url.QueryEscape(doc.String()),
			}

			req.AddHeader("X-Lovebeat", "1")

			res, err := req.Do()

			if err != nil {
				log.Error("Failed to post slackhook:%v:", err)
			}

			robots, err := ioutil.ReadAll(res.Body)
			res.Body.Close()

			//it returned a 200 so ignore any error here
			if err != nil {
				log.Error("OK:unreadable response:%v:", err)
			} else if res.StatusCode != http.StatusOK {
				log.Error("NOK:non-200:%d:", res.StatusCode)
			} else {
				log.Info("OK:response:%s:", string(robots))
			}

		}
	}

}

func NewSlackhookAlerter(cfg config.Config) Alerter {
	goreq.SetConnectTimeout(5 * time.Second)
	var q = make(chan slackhook, 100)
	var w = slackhookAlerter{cmds: q}
	go w.Worker(q, &cfg.Slackhook)
	return &w
}
