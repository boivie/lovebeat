package dashboard

import (
	"bytes"
	"fmt"
	"github.com/boivie/lovebeat-go/backend"
	"github.com/boivie/lovebeat-go/service"
	"github.com/gorilla/mux"
	"html/template"
	"io"
	"net/http"
	"strconv"
)

var client service.ServiceIf

func DashboardHandler(w http.ResponseWriter, r *http.Request) {
	tc := make(map[string]interface{})
	tc["services"] = client.GetServices("all")

	templates := template.Must(template.ParseFiles("templates/base.html", "templates/index.html"))
	if err := templates.Execute(w, tc); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func StatusHandler(c http.ResponseWriter, req *http.Request) {
	var buffer bytes.Buffer
	var services = client.GetServices("all")
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

func Register(rtr *mux.Router, client_ service.ServiceIf) {
	client = client_
	rtr.HandleFunc("/", DashboardHandler).Methods("GET")
	rtr.HandleFunc("/status", StatusHandler).Methods("GET")
}
