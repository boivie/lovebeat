package tcpapi

import (
	"bufio"
	"github.com/boivie/lovebeat-go/internal"
	"github.com/boivie/lovebeat-go/lineparser"
	"github.com/op/go-logging"
	"net"
)

var log = logging.MustGetLogger("lovebeat")
var ServiceCmdChan chan *internal.Cmd

func tcpHandle(c *net.TCPConn) {
	defer c.Close()
	r := bufio.NewReaderSize(c, 4096)
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		var buf = scanner.Bytes()
		for _, p := range lineparser.Parse(buf) {
			ServiceCmdChan <- p
		}
	}
}

func Listener(bindAddr string, channel chan *internal.Cmd) {
	ServiceCmdChan = channel
	address, _ := net.ResolveTCPAddr("tcp", bindAddr)
	log.Info("TCP listener running on %s", address)
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
