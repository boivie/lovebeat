// +build windows freebsd

package eventlog

import (
	"github.com/boivie/lovebeat/config"
	"github.com/boivie/lovebeat/eventbus"
	"os"
)

func Init(cfg config.Config, bus *eventbus.EventBus) {
	if len(cfg.Eventlog.Path) == 0 {
		return
	}
	log.Warning("Using compatibility mode for event logging")
	eventwriter, err := os.OpenFile(cfg.Eventlog.Path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, cfg.Eventlog.Mode)
	if err != nil {
		log.Errorf("Error opening event log for writing: %s", err)
	} else {
		log.Info("Logging events to %s", cfg.Eventlog.Path)
		evtlog := New(eventwriter)
		evtlog.Register(bus)
	}
}
