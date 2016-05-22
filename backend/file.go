package backend

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"github.com/boivie/lovebeat/config"
	"github.com/boivie/lovebeat/metrics"
	"github.com/boivie/lovebeat/model"
	"github.com/boivie/lovebeat/notify"
	"github.com/op/go-logging"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"io"
	"net/url"
)

const MAX_PENDING_WRITES = 1000

var (
	RECORD_SERVICE = []byte("SERV\t")
	RECORD_ALARM   = []byte("ALARM\t")
	NEWLINE        = []byte("\n")
)

var log = logging.MustGetLogger("lovebeat")

type FileBackend struct {
	cfg      *config.ConfigDatabase
	q        chan update
	sync     chan chan bool
	services map[string]*model.Service
	alarms   map[string]*model.Alarm
}

func (f FileBackend) Sync() {
	reply := make(chan bool)
	f.sync <- reply
	<-reply
}

func (f FileBackend) SaveService(service *model.Service) {
	f.q <- update{setService: service}
}

func (f FileBackend) SaveAlarm(alarm *model.Alarm) {
	f.q <- update{setAlarm: alarm}
}

func (r FileBackend) LoadServices() []*model.Service {
	v := make([]*model.Service, len(r.services))
	idx := 0
	for _, value := range r.services {
		v[idx] = value
		idx++
	}
	return v
}

func (r FileBackend) LoadAlarms() []*model.Alarm {
	v := make([]*model.Alarm, len(r.alarms))
	idx := 0
	for _, value := range r.alarms {
		v[idx] = value
		idx++
	}
	return v
}

func (f FileBackend) DeleteService(name string) {
	f.q <- update{deleteService: name}
}

func (f FileBackend) DeleteAlarm(name string) {
	f.q <- update{deleteAlarm: name}
}

func (f FileBackend) readAll() {
	s := f.cfg.Filename
	fi, err := os.Open(s)
	if err != nil {
		log.Errorf("Couldn't open '%s'\n", s)
		return
	}
	gz, err := gzip.NewReader(fi)
	if err != nil {
		log.Errorf("Couldn't read from '%s'\n", s)
		return
	}

	buf := bufio.NewReader(gz)
	for {
		line, err := buf.ReadBytes('\n')
		if err != nil {
			break
		}
		if bytes.HasPrefix(line, RECORD_SERVICE) {
			service := &model.Service{}
			json.Unmarshal(line[5:], &service)
			f.services[service.Name] = service
		} else if bytes.HasPrefix(line, RECORD_ALARM) {
			alarm := &model.Alarm{}
			json.Unmarshal(line[5:], &alarm)
			f.alarms[alarm.Name] = alarm
		} else {
			log.Infof("Found unexpected line in database - skipping")
		}
	}
	log.Infof("Loaded %d services and %d alarms from '%s'",
		len(f.services), len(f.alarms), s)
}

func (f FileBackend) saveAll(counters metrics.Metrics) {
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
	for _, alarm := range f.alarms {
		b, _ := json.Marshal(alarm)
		gz.Write(RECORD_ALARM)
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
	log.Debugf("Saved %d items in %d ms", len(f.services)+len(f.alarms),
		duration.Nanoseconds()/1000000)

	f.uploadToRemote()

	counters.IncCounter("db.save.count")
	counters.SetGauge("db.save.duration", int(duration.Nanoseconds()/1000000))
	counters.SetGauge("service.count", len(f.services))
	counters.SetGauge("alarm.count", len(f.alarms))
}

type update struct {
	setService    *model.Service
	setAlarm      *model.Alarm
	deleteService string
	deleteAlarm   string
}

func (f FileBackend) downloadFromRemote() {
	if f.cfg.RemoteS3Url != "" && f.cfg.RemoteS3Region != "" {
		log.Infof("Fetching database from '%s' (region '%s')", f.cfg.RemoteS3Url, f.cfg.RemoteS3Region)
		parsed, err := url.Parse(f.cfg.RemoteS3Url)
		if err != nil {
			log.Panicf("Failed to parse S3 url: %v", err)
		}
		svc := s3.New(session.New(), &aws.Config{Region: aws.String(f.cfg.RemoteS3Region)})
		resp, err := svc.GetObject(&s3.GetObjectInput{
			Bucket: aws.String(parsed.Host),
			Key:    aws.String(parsed.Path[1:]),
		})
		if err != nil {
			awse, ok := err.(awserr.Error)
			if !ok {
				log.Panicf("Unknown error getting database from S3: %v", err)
			}
			if awse.Code() == "NoSuchKey" {
				// This is expected. Just start from an empty database
				log.Warning("Couldn't find an initial database on S3 - starting with an empty one")
				return
			}
			log.Panicf("Unknown AWS error: %v", awse)
		}
		defer resp.Body.Close()

		s := f.cfg.Filename + ".new"
		fi, err := os.OpenFile(s, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0640)
		if err != nil {
			log.Panicf("Failed to create database from S3: %v", err)
		}
		_, err = io.Copy(fi, resp.Body)
		if err != nil {
			log.Panicf("Failed to read from S3: %v", err)
		}
		fi.Close()
		if err = os.Rename(s, f.cfg.Filename); err != nil {
			log.Panic("Failed to overwrite database")
		}
		log.Infof("Done fetching database from '%s'", f.cfg.RemoteS3Url)
	}
}

func (f FileBackend) uploadToRemote() {
	if f.cfg.RemoteS3Url != "" && f.cfg.RemoteS3Region != "" {
		log.Debugf("Uploading database to '%s' (region '%s')", f.cfg.RemoteS3Url, f.cfg.RemoteS3Region)
		parsed, err := url.Parse(f.cfg.RemoteS3Url)
		if err != nil {
			log.Errorf("Failed to parse S3 url: %v", err)
			return
		}
		svc := s3.New(session.New(), &aws.Config{Region: aws.String(f.cfg.RemoteS3Region)})

		fi, err := os.Open(f.cfg.Filename)
		if err != nil {
			log.Errorf("Failed to open  S3: %v", err)
			return
		}
		defer fi.Close()

		_, err = svc.PutObject(&s3.PutObjectInput{
			Bucket: aws.String(parsed.Host),
			Key:    aws.String(parsed.Path[1:]),
			Body:   fi,
		})
		if err != nil {
			log.Error("Unknown error uploading data to S3", err)
		}
		log.Infof("Uploaded database to '%s' (region '%s')", f.cfg.RemoteS3Url, f.cfg.RemoteS3Region)
	}
}

func (f FileBackend) monitor(counters metrics.Metrics, notifier notify.Notifier) {
	ticker := time.NewTicker(time.Duration(f.cfg.Interval) * time.Second)
	for {
		select {
		case <-ticker.C:
			f.saveAll(counters)
			notifier.Notify("save")
		case reply := <-f.sync:
			f.saveAll(counters)
			reply <- true
		case upd := <-f.q:
			if upd.setService != nil {
				f.services[upd.setService.Name] = upd.setService
			}
			if upd.deleteService != "" {
				delete(f.services, upd.deleteService)
			}
			if upd.setAlarm != nil {
				f.alarms[upd.setAlarm.Name] = upd.setAlarm
			}
			if upd.deleteAlarm != "" {
				delete(f.alarms, upd.deleteAlarm)
			}
		}
	}
}

func NewFileBackend(cfg *config.ConfigDatabase, m metrics.Metrics, notifier notify.Notifier) Backend {
	be := FileBackend{
		cfg:      cfg,
		q:        make(chan update, MAX_PENDING_WRITES),
		sync:     make(chan chan bool),
		services: make(map[string]*model.Service),
		alarms:   make(map[string]*model.Alarm),
	}
	be.downloadFromRemote()
	be.readAll()
	go be.monitor(m, notifier)
	return be
}
