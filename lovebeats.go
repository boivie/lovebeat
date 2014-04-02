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
	"github.com/hoisie/redis"
)

var signalchan chan os.Signal
var client redis.Client

const (
	VERSION                 = "0.1.0"
	MAX_UNPROCESSED_PACKETS = 1000
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
	WarningTimeout int64
	ErrorTimeout   int64
	State          string
	Status         string
}

var (
	In       = make(chan *Cmd, MAX_UNPROCESSED_PACKETS)
)

func now() int64 { return time.Now().Unix() }

func getOrCreate(name string) (*Service, *Service) {
	service := &Service{
		Name: name,
		LastValue: -1,
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

func getExpiry(service *Service, timeout int64) int64 {
	if timeout <= 0 {
		return 0
	}
	return service.LastBeat + timeout
}

func service_get_next_expiry(service *Service, ts int64) int64 {
	var next int64 = 0
	var warningExpiry = getExpiry(service, service.WarningTimeout)
	var errorExpiry = getExpiry(service, service.ErrorTimeout)
	if warningExpiry > 0 && warningExpiry > ts && (next == 0 || warningExpiry < next) {
		next = warningExpiry
	}
	if errorExpiry > 0 && errorExpiry > ts && (next == 0 || errorExpiry < next) {
		next = errorExpiry
	}
	log.Printf("now: %d, warning: %d, error: %d, chosen: %d", ts, warningExpiry, errorExpiry, next)
	return next
}

func service_update_state(service *Service, ts int64) {
	service.State = STATE_OK
	var warningExpiry = getExpiry(service, service.WarningTimeout)
	var errorExpiry = getExpiry(service, service.ErrorTimeout)
	if warningExpiry > 0 && ts >= warningExpiry {
		service.State = STATE_WARNING
	}
	if errorExpiry > 0 && ts >= errorExpiry {
		service.State = STATE_ERROR
	}
}

func updateExpiry(service *Service, ts int64) {
	if service.State != STATE_PAUSED {
		if expiry := service_get_next_expiry(service, ts); expiry > 0 {
			client.Zadd("lb.expiry", []byte(service.Name), float64(expiry))
			return
		}
	}
	client.Zrem("lb.expiry", []byte(service.Name))
}

func service_save(service *Service, ref *Service) {
	if *service != *ref {
		log.Printf("service " + service.Name + ", " + ref.State + " -> " + service.State);
		b, _ := json.Marshal(service)
		client.Set("lb.service." + service.Name, b)
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
			log.Printf("TICK!")
			// get list of expired events
			if expired, err := client.Zrangebyscore("lb.expiry", 0, float64(now())); err == nil {
				for _, elem := range expired {
					var service, ref = getOrCreate(string(elem))
					service_update_state(service, ts);
					service_save(service, ref)
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
			}
			service_update_state(service, ts);
			service_save(service, ref)
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

func main() {
	flag.Parse()
	if *showVersion {
		fmt.Printf("statsdaemon v%s (built w/%s)\n", VERSION, runtime.Version())
		return
	}

	signalchan = make(chan os.Signal, 1)
	signal.Notify(signalchan, syscall.SIGTERM)

	go udpListener()
	monitor()
}
