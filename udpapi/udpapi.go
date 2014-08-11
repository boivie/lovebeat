package udpapi


import (
	"net"
	"github.com/op/go-logging"
	"github.com/boivie/lovebeat-go/lineparser"
	"github.com/boivie/lovebeat-go/internal"
)

var log = logging.MustGetLogger("lovebeat")
var ServiceCmdChan chan *internal.Cmd

const (
	MAX_UDP_PACKET_SIZE     = 512

)
func Listener(bindAddr string, channel chan *internal.Cmd) {
	ServiceCmdChan = channel
	address, _ := net.ResolveUDPAddr("udp", bindAddr)
	log.Info("UDP listener running on %s", address)
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

		for _, p := range lineparser.Parse(message[:n]) {
			ServiceCmdChan <- p
		}
	}
}
