package tcpapi

import (
	"bufio"
	"github.com/boivie/lovebeat/config"
	"github.com/boivie/lovebeat/lineparser"
	"github.com/boivie/lovebeat/service"
	"github.com/op/go-logging"
	"net"
)

var log = logging.MustGetLogger("lovebeat")

func tcpHandle(c *net.TCPConn, iface service.ServiceIf) {
	defer c.Close()
	r := bufio.NewReaderSize(c, 4096)
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		var buf = scanner.Bytes()
		lineparser.Execute(lineparser.Parse(buf), iface)
	}
}

func Listener(cfg *config.ConfigBind, iface service.ServiceIf) {
	address, _ := net.ResolveTCPAddr("tcp", cfg.Listen)
	log.Info("TCP listening on %s", address)
	listener, err := net.ListenTCP("tcp", address)
	if err != nil {
		log.Fatalf("ListenTCP - %s", err)
	}
	for {
		c, err := listener.AcceptTCP()
		if nil != err {
			log.Error("Error: %s", err)
			break
		}
		go tcpHandle(c, iface)
	}
}
