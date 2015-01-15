package lineparser

import (
	"bytes"
	"github.com/boivie/lovebeat-go/service"
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

var packetRegexp = regexp.MustCompile("^([^:]+)\\.(beat|warn|err):(-?[0-9]+)\\|(g|c|ms)(\\|@([0-9\\.]+))?\n?$")

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
			iface.ConfigureService(cmd.Name, cmd.Value, 0)
		case "err":
			iface.ConfigureService(cmd.Name, 0, cmd.Value)
		case "beat":
			iface.Beat(cmd.Name)
		}
	}
}
