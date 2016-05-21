// +build windows freebsd

package eventlog

import (
	"github.com/boivie/lovebeat/config"
	"io"
	"os"
)

func makeWriter(cfg config.Config) (io.Writer, error) {
	log.Warning("Using compatibility mode for event logging")
	w, e := os.OpenFile(cfg.Eventlog.Path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, cfg.Eventlog.Mode)
	return w, e
}
