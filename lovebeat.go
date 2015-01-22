package main

import (
	"flag"
	"fmt"
	"github.com/boivie/lovebeat-go/alert"
	"github.com/boivie/lovebeat-go/backend"
	"github.com/boivie/lovebeat-go/config"
	"github.com/boivie/lovebeat-go/dashboard"
	"github.com/boivie/lovebeat-go/httpapi"
	"github.com/boivie/lovebeat-go/metrics"
	"github.com/boivie/lovebeat-go/service"
	"github.com/boivie/lovebeat-go/tcpapi"
	"github.com/boivie/lovebeat-go/udpapi"
	"github.com/gorilla/mux"
	"github.com/op/go-logging"
	"log/syslog"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
)

var log = logging.MustGetLogger("lovebeat")

const (
	VERSION = "0.8.0"
)

var (
	debug       = flag.Bool("debug", false, "Enable debug logs")
	showVersion = flag.Bool("version", false, "Print version string")
	cfgFile     = flag.String("config", "/etc/lovebeat.cfg", "Configuration file")
	useSyslog   = flag.Bool("syslog", false, "Log to syslog instead of stderr")
)

var (
	signalchan = make(chan os.Signal, 1)
)

func signalHandler(be backend.Backend) {
	for {
		select {
		case sig := <-signalchan:
			fmt.Printf("!! Caught signal %d... shutting down\n", sig)
			be.Sync()
			return
		}
	}
}

func httpServer(cfg *config.ConfigBind, svcs *service.Services) {
	rtr := mux.NewRouter()
	httpapi.Register(rtr, svcs.GetClient())
	dashboard.Register(rtr, svcs.GetClient())
	http.Handle("/", rtr)
	log.Info("HTTP listening on %s\n", cfg.Listen)
	http.ListenAndServe(cfg.Listen, nil)
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

	if *debug {
		logging.SetLevel(logging.DEBUG, "lovebeat")
	} else {
		logging.SetLevel(logging.INFO, "lovebeat")
	}
	if *useSyslog {
		var backend, err = logging.NewSyslogBackendPriority("lovebeat", syslog.LOG_DAEMON)
		if err != nil {
			panic(err)
		}
		logging.SetBackend(logging.AddModuleLevel(backend))
	} else {
		var format = logging.MustStringFormatter("%{level} %{message}")
		logging.SetFormatter(format)
	}
	log.Debug("Debug logs enabled")

	if *showVersion {
		fmt.Printf("lovebeats v%s (built w/%s)\n", VERSION, runtime.Version())
		return
	}

	var cfg = config.ReadConfig(*cfgFile)

	var hostname = getHostname()
	log.Info("Lovebeat v%s started as host %s, PID %d", VERSION, hostname, os.Getpid())

	m := metrics.New(&cfg.Metrics)

	var be = backend.NewFileBackend(&cfg.Database, m)
	var alerters = []alert.Alerter{alert.NewMailAlerter(&cfg.Mail),
		alert.NewWebhooksAlerter()}
	var svcs = service.NewServices(be, alerters, m)

	signal.Notify(signalchan, syscall.SIGTERM)
	signal.Notify(signalchan, os.Interrupt)

	go svcs.Monitor()
	go httpServer(&cfg.Http, svcs)
	go udpapi.Listener(&cfg.Udp, svcs.GetClient())
	go tcpapi.Listener(&cfg.Tcp, svcs.GetClient())

	// Ensure that the 'all' view exists
	svcs.GetClient().CreateOrUpdateView("all", "", "", "")

	m.IncCounter("started.count")

	signalHandler(be)
}
