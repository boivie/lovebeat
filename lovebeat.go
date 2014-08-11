package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"regexp"
	"runtime"
	"syscall"
	"strconv"
	"time"
	"net/http"
	"github.com/gorilla/mux"
	"github.com/op/go-logging"
	"github.com/boivie/lovebeat-go/internal"
	"github.com/boivie/lovebeat-go/service"
	"github.com/boivie/lovebeat-go/dashboard"
	"github.com/boivie/lovebeat-go/httpapi"
)

var log = logging.MustGetLogger("lovebeat")

const (
	VERSION                 = "0.1.0"
	MAX_UNPROCESSED_PACKETS = 1000
	MAX_UDP_PACKET_SIZE     = 512
)

var (
	serviceAddress   = flag.String("address", ":8127", "UDP service address")
	expiryInterval   = flag.Int64("expiry-interval", 1, "Expiry interval (seconds)")
	debug            = flag.Bool("debug", false, "print statistics sent to graphite")
	showVersion      = flag.Bool("version", false, "print version string")
)

var (
	ServiceCmdChan    = make(chan *internal.Cmd, MAX_UNPROCESSED_PACKETS)
	ViewCmdChan         = make(chan *internal.ViewCmd, MAX_UNPROCESSED_PACKETS)
	signalchan chan os.Signal
)

func now() int64 { return time.Now().Unix() }

func monitor() {
	period := time.Duration(*expiryInterval) * time.Second
	ticker := time.NewTicker(period)
	for {
		select {
		case sig := <-signalchan:
			fmt.Printf("!! Caught signal %d... shutting down\n", sig)
			return
		case <-ticker.C:
			var ts = now()
			for _, s := range service.GetServices() {
				if (s.State == service.STATE_PAUSED || s.State == s.StateAt(ts)) {
					continue;
				}
				var ref = *s
				s.State = s.StateAt(ts)
				s.Save(&ref, ts)
				s.UpdateViews(ViewCmdChan)
			}
		case c := <-ViewCmdChan:
			var ts = now()
			switch c.Action {
			case service.ACTION_REFRESH_VIEW:
				log.Debug("Refresh view %s", c.View)
				var view = service.GetView(c.View)
				var ref = *view
				view.Refresh(ts)
				view.Save(&ref, ts);
			}
		case c := <-ServiceCmdChan:
			var ts = now()
			var s = service.GetService(c.Service)
			var ref = *s
			switch c.Action {
			case internal.ACTION_SET_WARN:
				s.WarningTimeout = int64(c.Value)
			case internal.ACTION_SET_ERR:
				s.ErrorTimeout = int64(c.Value)
			case internal.ACTION_BEAT:
				if c.Value > 0 {
					s.LastBeat = ts
					var diff = ts - ref.LastBeat
					s.Log("%d|beat|%d", ts, diff)
					log.Debug("Beat from %s", s.Name)
				}
			}
			s.State = s.StateAt(ts)
			s.Save(&ref, ts)
			s.UpdateViews(ViewCmdChan)
		}
	}
}

var packetRegexp = regexp.MustCompile("^([^:]+)\\.(beat|warn|err):(-?[0-9]+)\\|(g|c|ms)(\\|@([0-9\\.]+))?\n?$")

func parseMessage(data []byte) []*internal.Cmd {
	var output []*internal.Cmd
	for _, line := range bytes.Split(data, []byte("\n")) {
		if len(line) == 0 {
			continue
		}

		item := packetRegexp.FindSubmatch(line)
		if len(item) == 0 {
			continue
		}

		var value int
		modifier := string(item[4])
		switch modifier {
		case "c":
			var vali, err = strconv.ParseInt(string(item[3]), 10, 64)
			if err != nil {
				log.Error("failed to ParseInt %s - %s", item[3], err)
				continue
			}
			value = int(vali)
		default:
			var valu, err = strconv.ParseUint(string(item[3]), 10, 64)
			if err != nil {
				log.Error("failed to ParseUint %s - %s", item[3], err)
				continue
			}
			value = int(valu)
		}
		var action string
		switch string(item[2]) {
		case "warn":
			action = internal.ACTION_SET_WARN
		case "err":
			action = internal.ACTION_SET_ERR
		case "beat":
			action = internal.ACTION_BEAT
		}
		

		packet := &internal.Cmd{
			Action: action,
			Service: string(item[1]),
			Value:    value,
		}
		output = append(output, packet)
	}
	return output
}

func udpListener() {
	address, _ := net.ResolveUDPAddr("udp", *serviceAddress)
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

		for _, p := range parseMessage(message[:n]) {
			ServiceCmdChan <- p
		}
	}
}

func tcpHandle(c *net.TCPConn) {
	defer c.Close()
	r := bufio.NewReaderSize(c, 4096)
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		var buf = scanner.Bytes()
		for _, p := range parseMessage(buf) {
			ServiceCmdChan <- p
		}
	}
}

func tcpListener() {
	address, _ := net.ResolveTCPAddr("tcp", *serviceAddress)
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


func httpServer(port int16) {
	rtr := mux.NewRouter()
	httpapi.Register(rtr, ServiceCmdChan, ViewCmdChan)
	dashboard.Register(rtr)
	http.Handle("/", rtr)
	log.Info("HTTP server running on port %d\n", port)
        http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

func main() {
	flag.Parse()
	if *showVersion {
		fmt.Printf("lovebeats v%s (built w/%s)\n", VERSION, runtime.Version())
		return
	}

	service.Startup()

	signalchan = make(chan os.Signal, 1)
	signal.Notify(signalchan, syscall.SIGTERM)

	var format = logging.MustStringFormatter("%{level} %{message}")
	logging.SetFormatter(format)
	if *debug {
		logging.SetLevel(logging.DEBUG, "lovebeat")
	} else {
		logging.SetLevel(logging.INFO, "lovebeat")
	}
	log.Debug("Debug logs enabled")

	go httpServer(8080)
	go udpListener()
	go tcpListener()
	monitor()
}
