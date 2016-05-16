// +build !windows,!freebsd

package eventlog

import (
	"github.com/boivie/lovebeat/config"
	"github.com/boivie/lovebeat/eventbus"
	"github.com/mipearson/rfw"
)

func Init(cfg config.Config, bus *eventbus.EventBus) {
	if len(cfg.Eventlog.Path) == 0 {
		return
	}
	eventwriter, err := rfw.Open(cfg.Eventlog.Path, cfg.Eventlog.Mode)
	if err != nil {
		log.Errorf("Error opening event log for writing: %s", err)
	} else {
		log.Infof("Logging events to %s", cfg.Eventlog.Path)
		evtlog := New(eventwriter)
		evtlog.Register(bus)
	}
}
