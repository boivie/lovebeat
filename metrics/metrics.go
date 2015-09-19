package metrics

import (
	"fmt"
	"github.com/boivie/lovebeat/config"
	"github.com/op/go-logging"
	"net"
)

var (
	log = logging.MustGetLogger("lovebeat")
)

type Metrics interface {
	IncCounter(name string)
	SetGauge(name string, value int)
}

type UdpMetrics struct {
	con    *net.UDPConn
	prefix string
}

func (s *UdpMetrics) IncCounter(name string) {
	data := fmt.Sprintf("%s.%s:1|c\n", s.prefix, name)
	s.send([]byte(data))
}

func (s *UdpMetrics) SetGauge(name string, value int) {
	data := fmt.Sprintf("%s.%s:%d|g\n", s.prefix, name, value)
	s.send([]byte(data))
}

func (s *UdpMetrics) send(data []byte) {
	n, err := s.con.Write(data)
	if err != nil {
		log.Error("Failed to write metrics: %s", err)
	}
	if n == 0 {
		log.Error("Failed to write metrics")
	}
}

func NewUdpMetrics(addr, prefix string) (*UdpMetrics, error) {
	serverAddr, err := net.ResolveUDPAddr("udp", addr)
	con, err := net.DialUDP("udp", nil, serverAddr)

	client := &UdpMetrics{
		con:    con,
		prefix: prefix}

	return client, err
}

type nopMetrics struct {
}

func (s *nopMetrics) IncCounter(name string) {
}

func (s *nopMetrics) SetGauge(name string, val int) {
}

var NOP = nopMetrics{}

func NopMetrics() Metrics {
	return &NOP
}

func New(cfg *config.ConfigMetrics) Metrics {
	if cfg.Server != "" {
		m, err := NewUdpMetrics(cfg.Server, cfg.Prefix)
		if err == nil {
			log.Info("Reporting metrics to %s", cfg.Server)
			return m
		}
	}
	log.Info("No metrics reporting configured")
	return NopMetrics()
}
