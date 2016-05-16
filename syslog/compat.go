// +build windows

package eventlog

import (
	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("lovebeat")

func Init() {
	log.Warning("Syslog logging not available on Windows")
}
