package httpapi

import (
	"github.com/boivie/lovebeat-go/service"
	"github.com/gorilla/mux"
	"github.com/op/go-logging"
	"io"
	"net/http"
	"strconv"
	"time"
)

var (
	svcs   *service.Services
	client service.ServiceIf
)

func now() int64 { return time.Now().Unix() }

var log = logging.MustGetLogger("lovebeat")

func ServiceHandler(c http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	name := params["name"]

	var err = r.ParseForm()
	if err != nil {
		log.Error("error parsing form ", err)
		return
	}

	var errtmo, warntmo = r.FormValue("err-tmo"), r.FormValue("warn-tmo")

	client.Beat(name)

	if val, err := strconv.Atoi(errtmo); err == nil {
		client.SetErrorTimeout(name, val)
	}

	if val, err := strconv.Atoi(warntmo); err == nil {
		client.SetWarningTimeout(name, val)
	}

	c.Header().Add("Content-Type", "text/plain")
	c.Header().Add("Content-Length", "3")
	io.WriteString(c, "{}\n")
}

func DeleteServiceHandler(c http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	name := params["name"]

	client.DeleteService(name)

	c.Header().Add("Content-Type", "text/plain")
	c.Header().Add("Content-Length", "3")
	io.WriteString(c, "{}\n")
}

func CreateViewHandler(c http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	view_name := params["name"]
	var expr = r.FormValue("regexp")
	if expr == "" {
		log.Error("No regexp provided")
		return
	}

	client.CreateOrUpdateView(view_name, expr)

	c.Header().Add("Content-Type", "text/plain")
	c.Header().Add("Content-Length", "3")
	io.WriteString(c, "{}\n")
}

func DeleteViewHandler(c http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	name := params["name"]

	client.DeleteView(name)

	c.Header().Add("Content-Type", "text/plain")
	c.Header().Add("Content-Length", "3")
	io.WriteString(c, "{}\n")
}

func Register(rtr *mux.Router, services *service.Services) {
	svcs = services
	client = svcs.GetClient()
	rtr.HandleFunc("/api/service/{name:[a-z0-9.]+}", ServiceHandler).Methods("POST")
	rtr.HandleFunc("/api/service/{name:[a-z0-9.]+}", DeleteServiceHandler).Methods("DELETE")
	rtr.HandleFunc("/api/view/{name:[a-z0-9.]+}", CreateViewHandler).Methods("POST")
	rtr.HandleFunc("/api/view/{name:[a-z0-9.]+}", DeleteViewHandler).Methods("DELETE")
}
