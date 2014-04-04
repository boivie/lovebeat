package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
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
)

const (
	VERSION                 = "0.1.0"
	MAX_UNPROCESSED_PACKETS = 1000
	MAX_LOG_ENTRIES         = 1000
	MAX_UDP_PACKET_SIZE     = 512
)

var (
	serviceAddress   = flag.String("address", ":8125", "UDP service address")
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
	In       = make(chan *Cmd, MAX_UNPROCESSED_PACKETS)
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
			log.Printf("service %s, state %s -> %s",
				s.Name, ref.State, s.State)
			s.Log("%d|state|%s", ts, s.State)
		}
		if s.WarningTimeout != ref.WarningTimeout {
			log.Printf("service %s, warn %d -> %d",
				s.Name, ref.WarningTimeout, s.WarningTimeout)
			s.Log("%d|warn-tmo|%s", ts, ref.WarningTimeout)
		}
		if s.ErrorTimeout != ref.ErrorTimeout {
			log.Printf("service %s, err %d -> %d",
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
				}
			}
		case s := <-In:
			var ts = now()
			var service, ref = getOrCreate(s.Service)
			switch s.Action {
			case ACTION_SET_WARN:
				service.WarningTimeout = int64(s.Value)
			case ACTION_SET_ERR:
				service.ErrorTimeout = int64(s.Value)
			case ACTION_BEAT:
				service.LastBeat = ts
				var diff = ts - ref.LastBeat
				service.Log("%d|beat|%d", ts, diff)
			}
			service.UpdateState(ts)
			service.Save(ref, ts)
			updateExpiry(service, ts)
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
				log.Printf("ERROR: failed to ParseInt %s - %s", item[3], err)
				continue
			}
			value = int(vali)
		default:
			var valu, err = strconv.ParseUint(string(item[3]), 10, 64)
			if err != nil {
				log.Printf("ERROR: failed to ParseUint %s - %s", item[3], err)
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
	log.Printf("listening on %s", address)
	listener, err := net.ListenUDP("udp", address)
	if err != nil {
		log.Fatalf("ERROR: ListenUDP - %s", err)
	}
	defer listener.Close()

	message := make([]byte, MAX_UDP_PACKET_SIZE)
	for {
		n, remaddr, err := listener.ReadFromUDP(message)
		if err != nil {
			log.Printf("ERROR: reading UDP packet from %+v - %s", remaddr, err)
			continue
		}

		for _, p := range parseMessage(message[:n]) {
			In <- p
		}
	}
}

func DashboardHandler(c http.ResponseWriter, req *http.Request) {
	var buffer bytes.Buffer
	var services, _ = client.Smembers("lb.services.all")
	var errors, warnings = false, false
	for _, v := range services {
		var service, _ = getOrCreate(string(v))
		buffer.WriteString(fmt.Sprintf("%-40s%s\n", service.Name, service.State))
		if service.State == STATE_WARNING {
			warnings = true
		}
		if service.State == STATE_ERROR {
			errors = true
		}
	}
	buffer.WriteString(fmt.Sprintf("\nwarnings: %t\nerrors: %t\ngood: %t\n", warnings, errors, !warnings && !errors))
        body := buffer.String()
        c.Header().Add("Content-Type", "text/plain")
        c.Header().Add("Content-Length", strconv.Itoa(len(body)))
        io.WriteString(c, body)
}

func StatusHandler(c http.ResponseWriter, req *http.Request) {
	var buffer bytes.Buffer
	var services, _ = client.Smembers("lb.services.all")
	var errors, warnings = false, false
	for _, v := range services {
		var service, _ = getOrCreate(string(v))
		if service.State == STATE_WARNING {
			warnings = true
		}
		if service.State == STATE_ERROR {
			errors = true
		}
	}
	buffer.WriteString(fmt.Sprintf("warnings: %t\nerrors: %t\ngood: %t\n", warnings, errors, !warnings && !errors))
        body := buffer.String()
        c.Header().Add("Content-Type", "text/plain")
        c.Header().Add("Content-Length", strconv.Itoa(len(body)))
        io.WriteString(c, body)
}

func TriggerHandler(c http.ResponseWriter, r *http.Request) {
	var err = r.ParseForm()
	if err != nil {
		log.Print("error parsing form ", err)
		return
	}

	var errtmo, warntmo = r.FormValue("err-tmo"), r.FormValue("warn-tmo")

	In <- &Cmd{
		Action:  ACTION_BEAT,
		Service: string("path"),
		Value:   1,
	}

	
	if val, err := strconv.Atoi(errtmo); err == nil {
		In <- &Cmd{
			Action:  ACTION_SET_ERR,
			Service: string("path"),
			Value:   val,
		}
	}

	if val, err := strconv.Atoi(warntmo); err == nil {
		In <- &Cmd{
			Action:  ACTION_SET_WARN,
			Service: string("path"),
			Value:   val,
		}
	}


        c.Header().Add("Content-Type", "text/plain")
        c.Header().Add("Content-Length", "3")
        io.WriteString(c, "ok\n")
}

func httpServer() {
	http.Handle("/", http.HandlerFunc(DashboardHandler))
	http.Handle("/status", http.HandlerFunc(StatusHandler))
	http.Handle("/trigger", http.HandlerFunc(TriggerHandler))
        http.ListenAndServe(":8080", nil)
}

func main() {
	flag.Parse()
	if *showVersion {
		fmt.Printf("statsdaemon v%s (built w/%s)\n", VERSION, runtime.Version())
		return
	}

	signalchan = make(chan os.Signal, 1)
	signal.Notify(signalchan, syscall.SIGTERM)

	go httpServer()
	go udpListener()
	monitor()
}
