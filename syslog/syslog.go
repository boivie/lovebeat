// +build !windows

package eventlog

import (
	"github.com/op/go-logging"
	"log/syslog"
)

func Init() {
	backend, err := logging.NewSyslogBackendPriority("lovebeat", syslog.LOG_DAEMON)
	if err != nil {
		panic(err)
	}
	logging.SetBackend(logging.AddModuleLevel(backend))

}
