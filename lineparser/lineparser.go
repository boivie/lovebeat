package lineparser

import (
	"bytes"
	"github.com/boivie/lovebeat/service"
	"github.com/op/go-logging"
	"regexp"
	"strconv"
)

type LineCommand struct {
	Action string
	Name   string
	Value  int64
}

var log = logging.MustGetLogger("lovebeat")

var packetRegexp = regexp.MustCompile("^(" + service.ServiceNamePattern +
	")\\.(beat|warn|err|autobeat):(-?[0-9]+)\\|(g|c|ms)(\\|@([0-9\\.]+))?\n?$")

func Parse(data []byte) []LineCommand {
	var commands []LineCommand

	for _, line := range bytes.Split(data, []byte("\n")) {
		if len(line) == 0 {
			continue
		}

		item := packetRegexp.FindSubmatch(line)
		if len(item) == 0 {
			continue
		}

		var vali, err = strconv.ParseInt(string(item[3]), 10, 64)
		if err != nil {
			log.Error("failed to ParseInt %s - %s", item[3], err)
			continue
		}

		if vali > 0 {
			vali = vali * 1000
		}

		cmd := LineCommand{
			Action: string(item[2]),
			Name:   string(item[1]),
			Value:  int64(vali)}
		commands = append(commands, cmd)
	}
	return commands
}

func Execute(commands []LineCommand, iface service.ServiceIf) {
	for _, cmd := range commands {
		switch cmd.Action {
		case "warn":
			iface.UpdateService(cmd.Name, false, cmd.Value, 0)
		case "err":
			iface.UpdateService(cmd.Name, false, 0, cmd.Value)
		case "beat":
			iface.UpdateService(cmd.Name, true, 0, 0)
		case "autobeat":
			iface.UpdateService(cmd.Name, true, 0, service.TIMEOUT_AUTO)
		}
	}
}
