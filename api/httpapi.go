package api

import (
	"encoding/json"
	"github.com/boivie/lovebeat/model"
	"github.com/boivie/lovebeat/service"
	"github.com/gorilla/mux"
	"io"
	"net/http"
	"strconv"
)

func parseTimeout(tmo string) int64 {
	if tmo == "auto" {
		return model.TIMEOUT_AUTO
	} else {
		val, _ := strconv.Atoi(tmo)
		return int64(val) * 1000
	}
}

func replyJson(c http.ResponseWriter, js interface{}) {
	var encoded, _ = json.MarshalIndent(js, "", "  ")
	c.Header().Add("Content-Type", "application/json")
	c.Header().Add("Content-Length", strconv.Itoa(len(encoded)+1))
	c.Write(encoded)
	io.WriteString(c, "\n")
}

func ServiceHandler(c http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	name := params["name"]

	update := &service.Update{Ts: now(), Service: name, Beat: &service.Beat{}}

	// Timeout as query parameter
	if val, ok := r.URL.Query()["timeout"]; ok {
		update.SetTimeout = &service.SetTimeout{Timeout: parseTimeout(val[0])}
	}

	if r.Header.Get("Content-Type") == "application/json" {
		decoder := json.NewDecoder(r.Body)
		var t struct {
			Timeout *int64 `json:"timeout"`
		}
		err := decoder.Decode(&t)
		if err == nil && t.Timeout != nil {
			if *t.Timeout > 0 {
				update.SetTimeout = &service.SetTimeout{Timeout: *t.Timeout * 1000}
			} else {
				update.SetTimeout = &service.SetTimeout{Timeout: *t.Timeout}
			}
		}
	} else {
		var err = r.ParseForm()
		if err != nil {
			log.Errorf("error parsing form: %v", err)
			return
		}

		if r.FormValue("timeout") != "" {
			update.SetTimeout = &service.SetTimeout{Timeout: parseTimeout(r.FormValue("timeout"))}
		}
	}

	client.Update(update)

	c.Header().Add("Content-Type", "application/json")
	c.Header().Add("Content-Length", "3")
	io.WriteString(c, "{}\n")
}

func MuteServiceHandler(c http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	name := params["name"]

	client.Update(&service.Update{Ts: now(), Service: name, MuteService: &service.MuteService{Muted: true}})

	c.Header().Add("Content-Type", "application/json")
	c.Header().Add("Content-Length", "3")
	io.WriteString(c, "{}\n")
}

func UnmuteServiceHandler(c http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	name := params["name"]

	client.Update(&service.Update{Ts: now(), Service: name, MuteService: &service.MuteService{Muted: false}})

	c.Header().Add("Content-Type", "application/json")
	c.Header().Add("Content-Length", "3")
	io.WriteString(c, "{}\n")
}

func DeleteServiceHandler(c http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	name := params["name"]

	client.Update(&service.Update{Ts: now(), Service: name, DeleteService: &service.DeleteService{}})

	c.Header().Add("Content-Type", "application/json")
	c.Header().Add("Content-Length", "3")
	io.WriteString(c, "{}\n")
}

type JsonView struct {
	Name  string `json:"name"`
	State string `json:"state"`
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
	replyJson(c, ret)
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
		Name:  v.Name,
		State: v.State,
	}

	replyJson(c, js)
}

type JsonViewRef struct {
	Name string `json:"name"`
}

type JsonService struct {
	Name          string        `json:"name"`
	LastBeat      int64         `json:"last_beat"`
	LastBeatDelta int64         `json:"last_beat_delta"`
	Timeout       int64         `json:"timeout"`
	State         string        `json:"state"`
	Views         []JsonViewRef `json:"views,omitempty"`
	History       []int64       `json:"history,omitempty"`
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
			Name:          s.Name,
			LastBeat:      s.LastBeat,
			LastBeatDelta: now - s.LastBeat,
			Timeout:       s.Timeout,
			State:         s.State,
		}
		ret = append(ret, js)
	}
	replyJson(c, ret)
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
		Name:          s.Name,
		LastBeat:      s.LastBeat,
		LastBeatDelta: now - s.LastBeat,
		Timeout:       s.Timeout,
		State:         s.State,
		History:       s.BeatHistory,
	}
	replyJson(c, js)
}

func AddEndpoints(rtr *mux.Router) {
	rtr.HandleFunc("/api/services/", GetServicesHandler).Methods("GET")
	rtr.HandleFunc("/api/services/{name:"+service.ServiceNamePattern+"}", ServiceHandler).Methods("POST")
	rtr.HandleFunc("/api/services/{name:"+service.ServiceNamePattern+"}/mute", MuteServiceHandler).Methods("POST")
	rtr.HandleFunc("/api/services/{name:"+service.ServiceNamePattern+"}/unmute", UnmuteServiceHandler).Methods("POST")
	rtr.HandleFunc("/api/services/{name:"+service.ServiceNamePattern+"}", GetServiceHandler).Methods("GET")
	rtr.HandleFunc("/api/services/{name:"+service.ServiceNamePattern+"}", DeleteServiceHandler).Methods("DELETE")
	rtr.HandleFunc("/api/views/", GetViewsHandler).Methods("GET")
	rtr.HandleFunc("/api/views/{name:"+service.ServiceNamePattern+"}", GetViewHandler).Methods("GET")
}
