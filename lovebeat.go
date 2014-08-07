package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"net"
	"encoding/json"
	"os"
	"os/signal"
	"regexp"
	"runtime"
	"syscall"
	"strconv"
	"time"
	"io"
	"net/http"
	"github.com/hoisie/redis"
	"github.com/gorilla/mux"
	"github.com/op/go-logging"
	"html/template"
)

var log = logging.MustGetLogger("lovebeat")

const (
	VERSION                 = "0.1.0"
	MAX_UNPROCESSED_PACKETS = 1000
	MAX_LOG_ENTRIES         = 1000
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

const (
	STATE_PAUSED  = "paused"
	STATE_OK      = "ok"
	STATE_WARNING = "warning"
	STATE_ERROR   = "error"
)

type Cmd struct {
	Action   string
	Service  string
	Value    int
}

const (
	ACTION_REFRESH_VIEW = "refresh-view"
)

type ViewCmd struct {
	Action   string
	View     string
}

type View struct {
	Name           string
	State          string
	LastUpdated    int64
}

type Service struct {
	Name           string
	LastValue      int
	LastBeat       int64
	LastUpdated    int64
	WarningTimeout int64
	ErrorTimeout   int64
	State          string
}

var (
	ServiceCmdChan    = make(chan *Cmd, MAX_UNPROCESSED_PACKETS)
	ViewCmdChan         = make(chan *ViewCmd, MAX_UNPROCESSED_PACKETS)
	signalchan chan os.Signal
	client redis.Client
)

func now() int64 { return time.Now().Unix() }

func getOrCreate(name string) (*Service, *Service) {
	service := &Service{
		Name: name,
		LastValue: -1,
		LastBeat: -1,
		LastUpdated: -1,
		WarningTimeout: -1,
		ErrorTimeout: -1,
		State: STATE_PAUSED,
	}

	if data, err := client.Get("lb.service." + name); err == nil {
		json.Unmarshal(data, &service)
	}
	var ref = *service
	return service, &ref
}

func (s *Service)GetExpiry(timeout int64) int64 {
	if timeout <= 0 {
		return 0
	}
	return s.LastBeat + timeout
}

func (s *Service)GetNextExpiry(ts int64) int64 {
	var next int64 = 0
	var warningExpiry = s.GetExpiry(s.WarningTimeout)
	var errorExpiry = s.GetExpiry(s.ErrorTimeout)
	if warningExpiry > 0 && warningExpiry > ts && (next == 0 || warningExpiry < next) {
		next = warningExpiry
	}
	if errorExpiry > 0 && errorExpiry > ts && (next == 0 || errorExpiry < next) {
		next = errorExpiry
	}
	return next
}

func (s *Service) UpdateState(ts int64) {
	s.State = STATE_OK
	var warningExpiry = s.GetExpiry(s.WarningTimeout)
	var errorExpiry = s.GetExpiry(s.ErrorTimeout)
	if warningExpiry > 0 && ts >= warningExpiry {
		s.State = STATE_WARNING
	}
	if errorExpiry > 0 && ts >= errorExpiry {
		s.State = STATE_ERROR
	}
}

func (s *Service) Log(format string, args ...interface{}) {
	var key = "lb.service-log." + s.Name
	var log = fmt.Sprintf(format, args...)
	client.Lpush(key, []byte(log))
	client.Ltrim(key, 0, MAX_LOG_ENTRIES)
}

func updateExpiry(service *Service, ts int64) {
	if service.State != STATE_PAUSED {
		if expiry := service.GetNextExpiry(ts); expiry > 0 {
			client.Zadd("lb.expiry", []byte(service.Name), float64(expiry))
			return
		}
	}
	client.Zrem("lb.expiry", []byte(service.Name))
}

func (s *Service)Save(ref *Service, ts int64) {
	if *s != *ref {
		if s.State != ref.State {
			log.Info("SERVICE '%s', state %s -> %s",
				s.Name, ref.State, s.State)
			s.Log("%d|state|%s", ts, s.State)
		}
		if s.WarningTimeout != ref.WarningTimeout {
			log.Info("SERVICE '%s', warn %d -> %d",
				s.Name, ref.WarningTimeout, s.WarningTimeout)
			s.Log("%d|warn-tmo|%s", ts, ref.WarningTimeout)
		}
		if s.ErrorTimeout != ref.ErrorTimeout {
			log.Info("SERVICE '%s', err %d -> %d",
				s.Name, ref.ErrorTimeout, s.ErrorTimeout)
			s.Log("%d|err-tmo|%s", ts, ref.ErrorTimeout)
		}
		s.LastUpdated = ts
		b, _ := json.Marshal(s)
		client.Set("lb.service." + s.Name, b)
		if ref.LastUpdated < 0 {
			client.Sadd("lb.services.all", []byte(s.Name))
		}
	}
}

func getView(name string) (*View, *View) {
	view := &View{
		Name: name,
		State: STATE_OK,
	}

	if data, err := client.Get("lb.view." + name); err == nil {
		json.Unmarshal(data, &view)
	}
	var ref = *view
	return view, &ref
}

func (v *View) Refresh(ts int64) {
	var services, _ = client.Smembers("lb.view-contents." + v.Name)
	v.State = STATE_OK

	for _, serv := range services {
		var service, _ = getOrCreate(string(serv))
		if service.State == STATE_WARNING && v.State == STATE_OK  {
			v.State = STATE_WARNING
		} else if service.State == STATE_ERROR {
			v.State = STATE_ERROR
		}
	}
}

func (v *View) Log(format string, args ...interface{}) {
	var key = "lb.view-log." + v.Name
	var log = fmt.Sprintf(format, args...)
	client.Lpush(key, []byte(log))
	client.Ltrim(key, 0, MAX_LOG_ENTRIES)
}

func (v *View) Save(ref *View, ts int64) {
	if *v != *ref {
		if v.State != ref.State {
			log.Info("VIEW '%s', state %s -> %s",
				v.Name, ref.State, v.State)
			v.Log("%d|state|%s", ts, v.State)
		}
		v.LastUpdated = ts
		b, _ := json.Marshal(v)
		client.Set("lb.view." + v.Name, b)
	}
}

func (s *Service) triggerViews() {
	var views, _ = client.Smembers("lb.views.all")

	for _, view := range views {
		var view_name = string(view)
		var mbr, _ = client.Sismember("lb.view-contents." + view_name,
			[]byte(s.Name));
		if mbr {
			ViewCmdChan <- &ViewCmd{
				Action: ACTION_REFRESH_VIEW,
				View:   view_name,
			}
		}
	}
}


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
			if expired, err := client.Zrangebyscore("lb.expiry", 0, float64(now())); err == nil {
				for _, elem := range expired {
					var service, ref = getOrCreate(string(elem))
					service.UpdateState(ts)
					service.Save(ref, ts)
					updateExpiry(service, ts)
					service.triggerViews()
				}
			}
		case s := <-ViewCmdChan:
			var ts = now()
			switch s.Action {
			case ACTION_REFRESH_VIEW:
				var view, ref = getView(s.View)
				view.Refresh(ts)
				view.Save(ref, ts);
			}
		case s := <-ServiceCmdChan:
			var ts = now()
			var service, ref = getOrCreate(s.Service)
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
			updateExpiry(service, ts)
			service.triggerViews()
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
		STATE_WARNING: "  WARN   ",
		STATE_ERROR:   "      ERR",
		STATE_OK:      "OK       ",
	}[in]
}

func DashboardHandler(w http.ResponseWriter, r *http.Request) {
	var names, _ = client.Smembers("lb.services.all")
	services := make(map[string]*Service)

	for _, name := range names {
		var service, _ = getOrCreate(string(name))
		services[string(name)] = service
	}

	tc := make(map[string]interface{})
	tc["services"] = services

	templates := template.Must(template.ParseFiles("templates/base.html", "templates/index.html"))
	if err := templates.Execute(w, tc); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func StatusHandler(c http.ResponseWriter, req *http.Request) {
	var buffer bytes.Buffer
	var services, _ = client.Smembers("lb.services.all")
	var errors, warnings, ok = 0, 0, 0
	for _, v := range services {
		var service, _ = getOrCreate(string(v))
		if service.State == STATE_WARNING {
			warnings++
		} else if service.State == STATE_ERROR {
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

	// Match services
	var key = "lb.view-contents." + view_name
	client.Del(key)
	var services, _ = client.Smembers("lb.services.all")
	for _, name := range services {
		if re.Match(name) {
			client.Sadd(key, name)
		}
	}

	client.Sadd("lb.views.all", []byte(view_name))
	log.Info("VIEW '%s' created or updated.", view_name)
	ViewCmdChan <- &ViewCmd{
		Action: ACTION_REFRESH_VIEW,
		View:   view_name,
	}
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
