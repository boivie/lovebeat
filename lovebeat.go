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
	"io"
	"net/http"
	"github.com/gorilla/mux"
	"github.com/op/go-logging"
	"html/template"
	"github.com/boivie/lovebeat-go/service"
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

const (
	ACTION_SET_WARN = "set-warn"
	ACTION_SET_ERR = "set-err"
	ACTION_BEAT = "beat"
)

type Cmd struct {
	Action   string
	Service  string
	Value    int
}

var (
	ServiceCmdChan    = make(chan *Cmd, MAX_UNPROCESSED_PACKETS)
	ViewCmdChan         = make(chan *service.ViewCmd, MAX_UNPROCESSED_PACKETS)
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
			for _, s := range service.GetExpired(ts) {
				var ref = *s
				s.UpdateState(ts)
				s.Save(&ref, ts)
				s.UpdateExpiry(ts)
				s.UpdateViews(ViewCmdChan)
			}
		case s := <-ViewCmdChan:
			var ts = now()
			switch s.Action {
			case service.ACTION_REFRESH_VIEW:
				var view, ref = service.GetView(s.View)
				view.Refresh(ts)
				view.Save(ref, ts);
			}
		case s := <-ServiceCmdChan:
			var ts = now()
			var service, ref = service.GetOrCreate(s.Service)
			switch s.Action {
			case ACTION_SET_WARN:
				service.WarningTimeout = int64(s.Value)
			case ACTION_SET_ERR:
				service.ErrorTimeout = int64(s.Value)
			case ACTION_BEAT:
				if s.Value > 1 {
					service.ErrorTimeout = int64(s.Value)
				}
				service.LastBeat = ts
				var diff = ts - ref.LastBeat
				service.Log("%d|beat|%d", ts, diff)
				log.Debug("Beat from %s", s.Service)
			}
			service.UpdateState(ts)
			service.Save(ref, ts)
			service.UpdateExpiry(ts)
			service.UpdateViews(ViewCmdChan)
		}
	}
}

var packetRegexp = regexp.MustCompile("^([^:]+)\\.(beat|warn|err):(-?[0-9]+)\\|(g|c|ms)(\\|@([0-9\\.]+))?\n?$")

func parseMessage(data []byte) []*Cmd {
	var output []*Cmd
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
			action = ACTION_SET_WARN
		case "err":
			action = ACTION_SET_ERR
		case "beat":
			action = ACTION_BEAT
		}
		

		packet := &Cmd{
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

func DashboardState(in string) string {
	return map[string]string {
		service.STATE_WARNING: "  WARN   ",
		service.STATE_ERROR:   "      ERR",
		service.STATE_OK:      "OK       ",
	}[in]
}

func DashboardHandler(w http.ResponseWriter, r *http.Request) {
	var services = service.GetServices()

	tc := make(map[string]interface{})
	tc["services"] = services

	templates := template.Must(template.ParseFiles("templates/base.html", "templates/index.html"))
	if err := templates.Execute(w, tc); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func StatusHandler(c http.ResponseWriter, req *http.Request) {
	var buffer bytes.Buffer
	var services = service.GetServices()
	var errors, warnings, ok = 0, 0, 0
	for _, s := range services {
		if s.State == service.STATE_WARNING {
			warnings++
		} else if s.State == service.STATE_ERROR {
			errors++
		} else {
			ok++
		}
	}
	buffer.WriteString(fmt.Sprintf("num_ok %d\nnum_warning %d\nnum_error %d\n",
		ok, warnings, errors))
	buffer.WriteString(fmt.Sprintf("has_warning %t\nhas_error %t\ngood %t\n",
		warnings > 0, errors > 0, warnings == 0 && errors == 0))
        body := buffer.String()
        c.Header().Add("Content-Type", "text/plain")
        c.Header().Add("Content-Length", strconv.Itoa(len(body)))
        io.WriteString(c, body)
}

func TriggerHandler(c http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	name := params["name"]

	var err = r.ParseForm()
	if err != nil {
		log.Error("error parsing form ", err)
		return
	}

	var errtmo, warntmo = r.FormValue("err-tmo"), r.FormValue("warn-tmo")

	ServiceCmdChan <- &Cmd{
		Action:  ACTION_BEAT,
		Service: name,
		Value:   1,
	}

	
	if val, err := strconv.Atoi(errtmo); err == nil {
		ServiceCmdChan <- &Cmd{
			Action:  ACTION_SET_ERR,
			Service: name,
			Value:   val,
		}
	}

	if val, err := strconv.Atoi(warntmo); err == nil {
		ServiceCmdChan <- &Cmd{
			Action:  ACTION_SET_WARN,
			Service: name,
			Value:   val,
		}
	}


        c.Header().Add("Content-Type", "text/plain")
        c.Header().Add("Content-Length", "3")
        io.WriteString(c, "ok\n")
}

func CreateViewHandler(c http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	view_name := params["name"]
	var expr = r.FormValue("regexp")
	if expr == "" {
		log.Error("No regexp provided")
		return
	}

	var re, err = regexp.Compile(expr)
	if err != nil {
		log.Error("Invalid regexp: %s", err)
		return
	}

	service.CreateView(view_name, re, ViewCmdChan)
}

func httpServer(port int16) {
	rtr := mux.NewRouter()
	rtr.HandleFunc("/", DashboardHandler).Methods("GET")
	rtr.HandleFunc("/status", StatusHandler).Methods("GET")
	rtr.HandleFunc("/trigger/{name:[a-z0-9.]+}", TriggerHandler).Methods("POST")
	rtr.HandleFunc("/view/{name:[a-z0-9.]+}", CreateViewHandler).Methods("POST")
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
