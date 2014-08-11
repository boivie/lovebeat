package httpapi

import (
	"bytes"
	"fmt"
	"github.com/boivie/lovebeat-go/internal"
	"github.com/boivie/lovebeat-go/service"
	"github.com/gorilla/mux"
	"github.com/op/go-logging"
	"io"
	"net/http"
	"strconv"
	"time"
)

var (
	ServiceCmdChan chan *internal.Cmd
	ViewCmdChan    chan *internal.ViewCmd
)

func now() int64 { return time.Now().Unix() }

var log = logging.MustGetLogger("lovebeat")

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

	ServiceCmdChan <- &internal.Cmd{
		Action:  internal.ACTION_BEAT,
		Service: name,
		Value:   1,
	}

	if val, err := strconv.Atoi(errtmo); err == nil {
		ServiceCmdChan <- &internal.Cmd{
			Action:  internal.ACTION_SET_ERR,
			Service: name,
			Value:   val,
		}
	}

	if val, err := strconv.Atoi(warntmo); err == nil {
		ServiceCmdChan <- &internal.Cmd{
			Action:  internal.ACTION_SET_WARN,
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

	service.CreateView(view_name, expr, ViewCmdChan, now())
}

func Register(rtr *mux.Router, cmd_chan chan *internal.Cmd, view_chan chan *internal.ViewCmd) {
	ServiceCmdChan = cmd_chan
	ViewCmdChan = view_chan
	rtr.HandleFunc("/status", StatusHandler).Methods("GET")
	rtr.HandleFunc("/trigger/{name:[a-z0-9.]+}", TriggerHandler).Methods("POST")
	rtr.HandleFunc("/view/{name:[a-z0-9.]+}", CreateViewHandler).Methods("POST")
}
