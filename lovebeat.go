package main

import (
	"flag"
	"fmt"
	"github.com/boivie/lovebeat-go/backend"
	"github.com/boivie/lovebeat-go/dashboard"
	"github.com/boivie/lovebeat-go/httpapi"
	"github.com/boivie/lovebeat-go/internal"
	"github.com/boivie/lovebeat-go/service"
	"github.com/boivie/lovebeat-go/tcpapi"
	"github.com/boivie/lovebeat-go/udpapi"
	"github.com/gorilla/mux"
	"github.com/op/go-logging"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
)

var log = logging.MustGetLogger("lovebeat")

const (
	VERSION                 = "0.1.0"
	MAX_UNPROCESSED_PACKETS = 1000
)

var (
	udpAddr        = flag.String("udp", ":8127", "UDP service address")
	tcpAddr        = flag.String("tcp", ":8127", "TCP service address")
	expiryInterval = flag.Int64("expiry-interval", 1, "Expiry interval (seconds)")
	debug          = flag.Bool("debug", false, "print statistics sent to graphite")
	showVersion    = flag.Bool("version", false, "print version string")
)

var (
	ServiceCmdChan = make(chan *internal.Cmd, MAX_UNPROCESSED_PACKETS)
	ViewCmdChan    = make(chan *internal.ViewCmd, MAX_UNPROCESSED_PACKETS)
	signalchan     = make(chan os.Signal, 1)
)

func signalHandler() {
	for {
		select {
		case sig := <-signalchan:
			fmt.Printf("!! Caught signal %d... shutting down\n", sig)
			return
		}
	}
}

func httpServer(port int16, svcs *service.Services) {
	rtr := mux.NewRouter()
	httpapi.Register(rtr, ServiceCmdChan, ViewCmdChan, svcs)
	dashboard.Register(rtr, svcs)
	http.Handle("/", rtr)
	log.Info("HTTP server running on port %d\n", port)
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

func getHostname() string {
	var hostname, err = os.Hostname()
	if err != nil {
		return fmt.Sprintf("unknown_%d", os.Getpid())
	}
	return strings.Split(hostname, ".")[0]
}

func main() {
	flag.Parse()

	var format = logging.MustStringFormatter("%{level} %{message}")
	logging.SetFormatter(format)
	if *debug {
		logging.SetLevel(logging.DEBUG, "lovebeat")
	} else {
		logging.SetLevel(logging.INFO, "lovebeat")
	}
	log.Debug("Debug logs enabled")

	if *showVersion {
		fmt.Printf("lovebeats v%s (built w/%s)\n", VERSION, runtime.Version())
		return
	}
	var hostname = getHostname()
	log.Info("Lovebeat v%s started as host %s, PID %d", VERSION, hostname, os.Getpid())

	var be = backend.RedisBackend{}
	var svcs = &service.Services{}
	svcs.Startup(be)

	signal.Notify(signalchan, syscall.SIGTERM)

	go svcs.Monitor(ServiceCmdChan, ViewCmdChan, *expiryInterval)
	go httpServer(8080, svcs)
	go udpapi.Listener(*udpAddr, ServiceCmdChan)
	go tcpapi.Listener(*tcpAddr, ServiceCmdChan)

	log.Info("Ready to handle incoming connections")

	signalHandler()
}
