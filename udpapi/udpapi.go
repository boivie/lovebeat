package udpapi

import (
	"github.com/boivie/lovebeat/config"
	"github.com/boivie/lovebeat/lineparser"
	"github.com/boivie/lovebeat/service"
	"github.com/op/go-logging"
	"net"
)

var log = logging.MustGetLogger("lovebeat")

const (
	MAX_UDP_PACKET_SIZE = 512
)

func Listener(cfg *config.ConfigBind, iface service.ServiceIf) {
	address, _ := net.ResolveUDPAddr("udp", cfg.Listen)
	log.Info("UDP listening on %s", address)
	listener, err := net.ListenUDP("udp", address)
	if err != nil {
		log.Fatalf("ListenUDP - %s", err)
	}
	defer listener.Close()

	message := make([]byte, MAX_UDP_PACKET_SIZE)
	for {
		n, remaddr, err := listener.ReadFromUDP(message)
		if err != nil {
			log.Error("reading UDP packet from %+v - %s", remaddr, err)
			continue
		}

		lineparser.Execute(lineparser.Parse(message[:n]), iface)
	}
}
