package notify

import (
	"github.com/boivie/lovebeat/config"
	"github.com/franela/goreq"
	"github.com/op/go-logging"
	"net/url"
	"time"
)

var log = logging.MustGetLogger("lovebeat")

type Notifier interface {
	Notify(metric string)
}

type lovebeatNotifier struct {
	q chan string
}

func (f lovebeatNotifier) Notify(metric string) {
	select {
	case f.q <- metric:
	default:
		log.Error("Failed to notify heartbeats!")
	}
}

func sendLovebeat(url, name, metric string) {
	log.Debugf("Sending outgoing lovebeat '%s' to %s", metric, url)
	req := goreq.Request{
		Method:      "POST",
		Uri:         url + "/api/services/lovebeat." + name + "." + metric,
		Accept:      "application/json",
		ContentType: "application/json",
		UserAgent:   "Lovebeat",
		Timeout:     10 * time.Second,
		Body: struct {
			Time int64 `json:"timeout"`
		}{-2},
	}

	if _, err := req.Do(); err != nil {
		log.Errorf("Failed to post webhook: %v", err)
	}
}

func Init(name string, cfg []config.ConfigNotify) Notifier {
	var q = make(chan string, 10)
	be := lovebeatNotifier{q}
	go func() {
		for metric := range q {
			for _, c := range cfg {
				if c.Lovebeat != "" {
					if u, err := url.Parse(c.Lovebeat); err == nil {
						if u.Scheme == "http" || u.Scheme == "https" {
							sendLovebeat(c.Lovebeat, name, metric)
						}
					}
				}
			}
		}
	}()
	return be
}
