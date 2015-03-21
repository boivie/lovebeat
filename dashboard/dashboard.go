package dashboard

import (
	"bytes"
	"fmt"
	"github.com/boivie/lovebeat/backend"
	"github.com/boivie/lovebeat/service"
	assetfs "github.com/elazarl/go-bindata-assetfs"
	"github.com/gorilla/mux"
	"io"
	"net/http"
	"strconv"
)

var client service.ServiceIf

func StatusHandler(c http.ResponseWriter, req *http.Request) {
	viewName := "all"

	if val, ok := req.URL.Query()["view"]; ok {
		viewName = val[0]
	}

	var buffer bytes.Buffer
	var services = client.GetServices(viewName)
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

func RedirectHandler(c http.ResponseWriter, req *http.Request) {
	http.Redirect(c, req, "/dashboard.html", 301)
}

func Register(rtr *mux.Router, client_ service.ServiceIf) {
	client = client_
	rtr.HandleFunc("/", RedirectHandler).Methods("GET")
	rtr.HandleFunc("/status", StatusHandler).Methods("GET")
	rtr.PathPrefix("/").Handler(http.FileServer(
		&assetfs.AssetFS{Asset, AssetDir, "data/"}))
}
