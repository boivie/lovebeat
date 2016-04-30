package api

import (
	"bufio"
	"github.com/boivie/lovebeat/config"
	"net"
)

func tcpHandle(c *net.TCPConn) {
	defer c.Close()
	r := bufio.NewReaderSize(c, 4096)
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		var buf = scanner.Bytes()
		Execute(Parse(buf), client)
	}
}

func TcpListener(cfg *config.ConfigBind) {
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
		go tcpHandle(c)
	}
}
