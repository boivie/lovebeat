package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
	"net/http"
	"github.com/gorilla/mux"
	"github.com/op/go-logging"
	"github.com/boivie/lovebeat-go/internal"
	"github.com/boivie/lovebeat-go/service"
	"github.com/boivie/lovebeat-go/dashboard"
	"github.com/boivie/lovebeat-go/httpapi"
	"github.com/boivie/lovebeat-go/udpapi"
	"github.com/boivie/lovebeat-go/tcpapi"
)

var log = logging.MustGetLogger("lovebeat")

const (
	VERSION                 = "0.1.0"
	MAX_UNPROCESSED_PACKETS = 1000
)

var (
	udpAddr          = flag.String("udp", ":8127", "UDP service address")
	tcpAddr          = flag.String("tcp", ":8127", "TCP service address")
	expiryInterval   = flag.Int64("expiry-interval", 1, "Expiry interval (seconds)")
	debug            = flag.Bool("debug", false, "print statistics sent to graphite")
	showVersion      = flag.Bool("version", false, "print version string")
)

var (
	ServiceCmdChan    = make(chan *internal.Cmd, MAX_UNPROCESSED_PACKETS)
	ViewCmdChan         = make(chan *internal.ViewCmd, MAX_UNPROCESSED_PACKETS)
	signalchan = make(chan os.Signal, 1)

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
	log.Info("Lovebeat v%s started as PID %d", VERSION, os.Getpid())

	service.Startup()

	signal.Notify(signalchan, syscall.SIGTERM)

	go httpServer(8080)
	go udpapi.Listener(*udpAddr, ServiceCmdChan)
	go tcpapi.Listener(*tcpAddr, ServiceCmdChan)

	log.Info("Ready to handle incoming connections")

	monitor()
}
