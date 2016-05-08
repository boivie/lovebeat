package api

import (
	"bytes"
	"github.com/boivie/lovebeat/model"
	"github.com/boivie/lovebeat/service"
	"regexp"
	"strconv"
)

type LineCommand struct {
	Action string
	Name   string
	Value  int64
}

var packetRegexp = regexp.MustCompile("^(" + service.ServiceNamePattern +
	")\\.(beat|autobeat|timeout):(-?[0-9]+)\\|(g|c|ms)(\\|@([0-9\\.]+))?\n?$")

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
			log.Errorf("failed to ParseInt %s - %s", item[3], err)
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

func Execute(commands []LineCommand, iface service.Services) {
	for _, cmd := range commands {
		switch cmd.Action {
		case "timeout":
			client.Update(&service.Update{
				Ts:         now(),
				Service:    cmd.Name,
				SetTimeout: &service.SetTimeout{Timeout: cmd.Value}})
		case "beat":
			client.Update(&service.Update{
				Ts:      now(),
				Service: cmd.Name,
				Beat:    &service.Beat{}})
		case "autobeat":
			client.Update(&service.Update{
				Ts:         now(),
				Service:    cmd.Name,
				SetTimeout: &service.SetTimeout{Timeout: model.TIMEOUT_AUTO},
				Beat:       &service.Beat{}})
		}
	}
}
