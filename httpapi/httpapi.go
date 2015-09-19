package httpapi

import (
	"encoding/json"
	"github.com/boivie/lovebeat/service"
	"github.com/gorilla/mux"
	"github.com/op/go-logging"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"time"
)

var (
	client service.ServiceIf
)

func now() int64 { return int64(time.Now().UnixNano() / 1e6) }

var log = logging.MustGetLogger("lovebeat")

func parseTimeout(tmo string) int64 {
	if tmo == "auto" {
		return service.TIMEOUT_AUTO
	} else {
		val, _ := strconv.Atoi(tmo)
		return int64(val) * 1000
	}
}

func ServiceHandler(c http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	name := params["name"]

	var err = r.ParseForm()
	if err != nil {
		log.Error("error parsing form ", err)
		return
	}

	errval := parseTimeout(r.FormValue("err-tmo"))
	warnval := parseTimeout(r.FormValue("warn-tmo"))

	client.UpdateService(name, true, warnval, errval)

	c.Header().Add("Content-Type", "application/json")
	c.Header().Add("Content-Length", "3")
	io.WriteString(c, "{}\n")
}

func DeleteServiceHandler(c http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	name := params["name"]

	client.DeleteService(name)

	c.Header().Add("Content-Type", "application/json")
	c.Header().Add("Content-Length", "3")
	io.WriteString(c, "{}\n")
}

type JsonView struct {
	Name   string `json:"name"`
	State  string `json:"state"`
	Regexp string `json:"regexp,omitempty"`
}

func GetViewsHandler(c http.ResponseWriter, r *http.Request) {
	var ret = make([]JsonView, 0)
	for _, v := range client.GetViews() {
		js := JsonView{
			Name:  v.Name,
			State: v.State,
		}
		ret = append(ret, js)
	}
	var encoded, _ = json.MarshalIndent(ret, "", "  ")

	c.Header().Add("Content-Type", "application/json")
	c.Header().Add("Content-Length", strconv.Itoa(len(encoded)+1))
	c.Write(encoded)
	io.WriteString(c, "\n")
}

func GetViewHandler(c http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	name := params["name"]
	v := client.GetView(name)
	if v == nil {
		http.NotFound(c, r)
		return
	}
	js := JsonView{
		Name:   v.Name,
		State:  v.State,
		Regexp: v.Regexp,
	}

	var encoded, _ = json.MarshalIndent(js, "", "  ")

	c.Header().Add("Content-Type", "application/json")
	c.Header().Add("Content-Length", strconv.Itoa(len(encoded)+1))
	c.Write(encoded)
	io.WriteString(c, "\n")
}

type JsonViewRef struct {
	Name string `json:"name"`
}

type JsonService struct {
	Name           string        `json:"name"`
	LastBeat       int64         `json:"last_beat"`
	LastBeatDelta  int64         `json:"last_beat_delta"`
	WarningTimeout int64         `json:"warning_timeout"`
	ErrorTimeout   int64         `json:"error_timeout"`
	State          string        `json:"state"`
	Views          []JsonViewRef `json:"views,omitempty"`
	History        []int64       `json:"history,omitempty"`
}

func GetServicesHandler(c http.ResponseWriter, r *http.Request) {
	viewName := "all"

	if val, ok := r.URL.Query()["view"]; ok {
		viewName = val[0]
	}
	var now = now()
	var ret = make([]JsonService, 0)
	for _, s := range client.GetServices(viewName) {
		js := JsonService{
			Name:           s.Name,
			LastBeat:       s.LastBeat,
			LastBeatDelta:  now - s.LastBeat,
			WarningTimeout: s.WarningTimeout,
			ErrorTimeout:   s.ErrorTimeout,
			State:          s.State,
		}
		ret = append(ret, js)
	}
	var encoded, _ = json.MarshalIndent(ret, "", "  ")

	c.Header().Add("Content-Type", "application/json")
	c.Header().Add("Content-Length", strconv.Itoa(len(encoded)+1))
	c.Write(encoded)
	io.WriteString(c, "\n")
}

func GetServiceHandler(c http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	name := params["name"]

	var now = now()
	var s = client.GetService(name)
	if s == nil {
		http.NotFound(c, r)
		return
	}

	js := JsonService{
		Name:           s.Name,
		LastBeat:       s.LastBeat,
		LastBeatDelta:  now - s.LastBeat,
		WarningTimeout: s.WarningTimeout,
		ErrorTimeout:   s.ErrorTimeout,
		State:          s.State,
		History:        s.BeatHistory,
	}

	if _, ok := r.URL.Query()["details"]; ok {
		js.Views = make([]JsonViewRef, 0)
		for _, view := range client.GetViews() {
			matched, err := regexp.MatchString(view.Regexp, s.Name)
			if err == nil && matched {
				js.Views = append(js.Views, JsonViewRef{Name: view.Name})
			}
		}
	}
	var encoded, _ = json.MarshalIndent(js, "", "  ")

	c.Header().Add("Content-Type", "application/json")
	c.Header().Add("Content-Length", strconv.Itoa(len(encoded)+1))
	c.Write(encoded)
	io.WriteString(c, "\n")
}

func Register(rtr *mux.Router, client_ service.ServiceIf) {
	client = client_
	rtr.HandleFunc("/api/services/",
		GetServicesHandler).Methods("GET")
	rtr.HandleFunc("/api/services/{name:"+service.ServiceNamePattern+"}",
		ServiceHandler).Methods("POST")
	rtr.HandleFunc("/api/services/{name:"+service.ServiceNamePattern+"}",
		GetServiceHandler).Methods("GET")
	rtr.HandleFunc("/api/services/{name:"+service.ServiceNamePattern+"}",
		DeleteServiceHandler).Methods("DELETE")
	rtr.HandleFunc("/api/views/",
		GetViewsHandler).Methods("GET")
	rtr.HandleFunc("/api/views/{name:"+service.ServiceNamePattern+"}",
		GetViewHandler).Methods("GET")
}
