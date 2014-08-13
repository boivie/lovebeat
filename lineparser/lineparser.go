package lineparser

import (
	"bytes"
	"github.com/boivie/lovebeat-go/service"
	"github.com/op/go-logging"
	"regexp"
	"strconv"
)

var log = logging.MustGetLogger("lovebeat")

var packetRegexp = regexp.MustCompile("^([^:]+)\\.(beat|warn|err):(-?[0-9]+)\\|(g|c|ms)(\\|@([0-9\\.]+))?\n?$")

func Parse(data []byte, iface service.ServiceIf) {
	for _, line := range bytes.Split(data, []byte("\n")) {
		if len(line) == 0 {
			continue
		}

		item := packetRegexp.FindSubmatch(line)
		if len(item) == 0 {
			continue
		}

		var value int
		modifier := string(item[4])
		switch modifier {
		case "c":
			var vali, err = strconv.ParseInt(string(item[3]), 10, 64)
			if err != nil {
				log.Error("failed to ParseInt %s - %s", item[3], err)
				continue
			}
			value = int(vali)
		default:
			var valu, err = strconv.ParseUint(string(item[3]), 10, 64)
			if err != nil {
				log.Error("failed to ParseUint %s - %s", item[3], err)
				continue
			}
			value = int(valu)
		}
		var name = string(item[1])
		switch string(item[2]) {
		case "warn":
			iface.SetWarningTimeout(name, value)
		case "err":
			iface.SetErrorTimeout(name, value)
		case "beat":
			iface.Beat(name)
		}
	}
}
