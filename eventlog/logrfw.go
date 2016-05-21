// +build !windows,!freebsd

package eventlog

import (
	"github.com/boivie/lovebeat/config"
	"github.com/mipearson/rfw"
	"io"
)

func makeWriter(cfg config.Config) (io.Writer, error) {
	w, e := rfw.Open(cfg.Eventlog.Path, cfg.Eventlog.Mode)
	return w, e
}
