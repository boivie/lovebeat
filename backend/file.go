package backend

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"github.com/boivie/lovebeat/config"
	"github.com/boivie/lovebeat/metrics"
	"github.com/op/go-logging"
	"os"
	"time"
)

const (
	MAX_PENDING_WRITES = 1000
)

var (
	RECORD_SERVICE = []byte("SERV\t")
	RECORD_VIEW    = []byte("VIEW\t")
	NEWLINE        = []byte("\n")
)

var (
	log      = logging.MustGetLogger("lovebeat")
	counters = metrics.NopMetrics()
)

type FileBackend struct {
	cfg      *config.ConfigDatabase
	q        chan update
	sync     chan chan bool
	services map[string]*StoredService
	views    map[string]*StoredView
}

func (f FileBackend) Sync() {
	reply := make(chan bool)
	f.sync <- reply
	<-reply
}

func (f FileBackend) SaveService(service *StoredService) {
	f.q <- update{setService: service}
}

func (f FileBackend) SaveView(view *StoredView) {
	f.q <- update{setView: view}
}

func (r FileBackend) LoadServices() []*StoredService {
	v := make([]*StoredService, len(r.services))
	idx := 0
	for _, value := range r.services {
		v[idx] = value
		idx++
	}
	return v
}

func (r FileBackend) LoadViews() []*StoredView {
	v := make([]*StoredView, len(r.views))
	idx := 0
	for _, value := range r.views {
		v[idx] = value
		idx++
	}
	return v
}

func (f FileBackend) DeleteService(name string) {
	f.q <- update{deleteService: name}
}

func (f FileBackend) DeleteView(name string) {
	f.q <- update{deleteView: name}
}

func (f FileBackend) loadService(data []byte) {
	service := &StoredService{}
	json.Unmarshal(data, &service)
	f.services[service.Name] = service
}

func (f FileBackend) loadView(data []byte) {
	view := &StoredView{}
	json.Unmarshal(data, &view)
	f.views[view.Name] = view
}

func (f FileBackend) readAll() {
	s := f.cfg.Filename
	fi, err := os.Open(s)
	if err != nil {
		log.Error("Couldn't open '%s'\n", s)
		return
	}
	gz, err := gzip.NewReader(fi)
	if err != nil {
		log.Error("Couldn't read from '%s'\n", s)
		return
	}

	buf := bufio.NewReader(gz)
	for {
		line, err := buf.ReadBytes('\n')
		if err != nil {
			break
		}
		if bytes.HasPrefix(line, RECORD_SERVICE) {
			f.loadService(line[5:])
		} else if bytes.HasPrefix(line, RECORD_VIEW) {
			f.loadView(line[5:])
		} else {
			log.Info("Found unexpected line in database - skipping")
		}
	}
	log.Info("Loaded %d services and %d views from '%s'",
		len(f.services), len(f.views), s)
}

func (f FileBackend) saveAll() {
	start := time.Now()
	s := f.cfg.Filename + ".new"
	fi, err := os.OpenFile(s, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0640)
	if err != nil {
		log.Error("Error creating file\n")
		return
	}
	gz := gzip.NewWriter(fi)

	for _, service := range f.services {
		b, _ := json.Marshal(service)
		gz.Write(RECORD_SERVICE)
		gz.Write(b)
		gz.Write(NEWLINE)
	}
	for _, view := range f.views {
		b, _ := json.Marshal(view)
		gz.Write(RECORD_VIEW)
		gz.Write(b)
		gz.Write(NEWLINE)
	}
	gz.Close()
	fi.Close()
	if err = os.Rename(s, f.cfg.Filename); err != nil {
		log.Error("Failed to overwrite database")
		return
	}
	duration := time.Since(start)
	log.Debug("Saved %d items in %d ms", len(f.services)+len(f.views),
		duration.Nanoseconds()/1000000)

	counters.IncCounter("db.save.count")
	counters.SetGauge("db.save.duration", int(duration.Nanoseconds()/1000000))
	counters.SetGauge("service.count", len(f.services))
	counters.SetGauge("view.count", len(f.views))
}

type update struct {
	setService    *StoredService
	setView       *StoredView
	deleteService string
	deleteView    string
}

func (f FileBackend) fileSaver() {
	period := time.Duration(f.cfg.Interval) * time.Second
	ticker := time.NewTicker(period)
	for {
		select {
		case <-ticker.C:
			f.saveAll()
		case reply := <-f.sync:
			f.saveAll()
			reply <- true
		case upd := <-f.q:
			if upd.setService != nil {
				f.services[upd.setService.Name] = upd.setService
			}
			if upd.deleteService != "" {
				delete(f.services, upd.deleteService)
			}
			if upd.setView != nil {
				f.views[upd.setView.Name] = upd.setView
			}
			if upd.deleteView != "" {
				delete(f.views, upd.deleteView)
			}
		}
	}
}

func NewFileBackend(cfg *config.ConfigDatabase, m metrics.Metrics) Backend {
	counters = m
	var q = make(chan update, MAX_PENDING_WRITES)
	be := FileBackend{
		cfg:      cfg,
		q:        q,
		sync:     make(chan chan bool),
		services: make(map[string]*StoredService),
		views:    make(map[string]*StoredView),
	}
	be.readAll()
	go be.fileSaver()
	return be
}
