package api

import (
	"github.com/boivie/lovebeat/config"
	"net"
)

const (
	MAX_UDP_PACKET_SIZE = 512
)

func UdpListener(cfg *config.ConfigBind) {
	address, _ := net.ResolveUDPAddr("udp", cfg.Listen)
	log.Infof("UDP listening on %s", address)
	listener, err := net.ListenUDP("udp", address)
	if err != nil {
		log.Fatalf("ListenUDP - %s", err)
	}
	defer listener.Close()

	message := make([]byte, MAX_UDP_PACKET_SIZE)
	for {
		n, remaddr, err := listener.ReadFromUDP(message)
		if err != nil {
			log.Errorf("reading UDP packet from %+v - %s", remaddr, err)
			continue
		}

		Execute(Parse(message[:n]), client)
	}
}
