package httpapi

import (
	"bytes"
	"fmt"
	"github.com/boivie/lovebeat-go/backend"
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

func StatusHandler(c http.ResponseWriter, req *http.Request) {
	var buffer bytes.Buffer
	var services = svcs.GetServices()
	var errors, warnings, ok = 0, 0, 0
	for _, s := range services {
		if s.State == backend.STATE_WARNING {
			warnings++
		} else if s.State == backend.STATE_ERROR {
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
	io.WriteString(c, "ok\n")
}

func DeleteServiceHandler(c http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	name := params["name"]

	client.DeleteService(name)

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

	client.CreateOrUpdateView(view_name, expr)
}

func Register(rtr *mux.Router, services *service.Services) {
	svcs = services
	client = svcs.GetClient()
	rtr.HandleFunc("/status", StatusHandler).Methods("GET")
	rtr.HandleFunc("/service/{name:[a-z0-9.]+}", ServiceHandler).Methods("POST")
	rtr.HandleFunc("/service/{name:[a-z0-9.]+}", DeleteServiceHandler).Methods("DELETE")
	rtr.HandleFunc("/view/{name:[a-z0-9.]+}", CreateViewHandler).Methods("POST")
}
