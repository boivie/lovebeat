package main // import "github.com/boivie/lovebeat"

import (
	"flag"
	"fmt"
	"github.com/boivie/lovebeat/alert"
	"github.com/boivie/lovebeat/algorithms"
	"github.com/boivie/lovebeat/backend"
	"github.com/boivie/lovebeat/config"
	"github.com/boivie/lovebeat/dashboard"
	"github.com/boivie/lovebeat/eventbus"
	"github.com/boivie/lovebeat/eventlog"
	"github.com/boivie/lovebeat/api"
	"github.com/boivie/lovebeat/metrics"
	"github.com/boivie/lovebeat/service"
	"github.com/boivie/lovebeat/stream"
	"github.com/gorilla/mux"
	"github.com/mipearson/rfw"
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

var (
	debug       = flag.Bool("debug", false, "Enable debug logs")
	showVersion = flag.Bool("version", false, "Print version string")
	cfgFile     = flag.String("config", "/etc/lovebeat.cfg", "Configuration file")
	cfgDir      = flag.String("config-dir", "/etc/lovebeat.conf.d", "Configuration directory")
	useSyslog   = flag.Bool("syslog", false, "Log to syslog instead of stderr")
	validate    = flag.Bool("validate-auto", false, "Evaluate auto-algorithm")
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

func httpServer(cfg *config.ConfigBind, svcs *service.Services, bus *eventbus.EventBus) {
	rtr := mux.NewRouter()
	api.AddEndpoints(rtr)
	stream.Register(rtr, bus)
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

func setUpEventlog(cfg config.Config, bus *eventbus.EventBus) {
	if len(cfg.Eventlog.Path) == 0 {
		return
	}
	eventwriter, err := rfw.Open(cfg.Eventlog.Path, cfg.Eventlog.Mode)
	if err != nil {
		log.Error("Error opening event log for writing: %s", err)
	} else {
		log.Info("Logging events to %s", cfg.Eventlog.Path)
		evtlog := eventlog.New(eventwriter)
		evtlog.Register(bus)
	}
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

	if *validate {
		fmt.Fprintf(os.Stderr, "Validating auto-algorithm from stdin\n")
		algorithms.Validate()
		return
	}

	var hostname = getHostname()
	wd, _ := os.Getwd()
	log.Info("Lovebeat v%s started as host %s, PID %d, running from %s", VERSION, hostname, os.Getpid(), wd)

	var cfg = config.ReadConfig(*cfgFile, *cfgDir)
	bus := eventbus.New()

	setUpEventlog(cfg, bus)

	m := metrics.New(&cfg.Metrics)

	service.RegisterMetrics(bus, m)
	var be = backend.NewFileBackend(&cfg.Database, m)
	var svcs = service.NewServices(be, bus)

	alert.RegisterAlerters(bus, cfg)

	signal.Notify(signalchan, syscall.SIGTERM)
	signal.Notify(signalchan, os.Interrupt)

	api.Init(svcs.GetClient())

	go svcs.Monitor(cfg)
	go httpServer(&cfg.Http, svcs, bus)
	go api.UdpListener(&cfg.Udp)
	go api.TcpListener(&cfg.Tcp)

	m.IncCounter("started.count")

	signalHandler(be)
}
