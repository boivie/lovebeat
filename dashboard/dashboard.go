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

var svcs *service.Services

func DashboardHandler(w http.ResponseWriter, r *http.Request) {
	var services = svcs.GetServices()

	tc := make(map[string]interface{})
	tc["services"] = services

	templates := template.Must(template.ParseFiles("templates/base.html", "templates/index.html"))
	if err := templates.Execute(w, tc); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

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

func Register(rtr *mux.Router, services *service.Services) {
	svcs = services
	rtr.HandleFunc("/", DashboardHandler).Methods("GET")
	rtr.HandleFunc("/status", StatusHandler).Methods("GET")
}
