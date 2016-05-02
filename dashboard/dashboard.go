package dashboard

import (
	"bytes"
	"fmt"
	"github.com/boivie/lovebeat/model"
	"github.com/boivie/lovebeat/service"
	assetfs "github.com/elazarl/go-bindata-assetfs"
	"github.com/gorilla/mux"
	"io"
	"net/http"
	"strconv"
	"encoding/json"
)

var client service.ServiceIf
var rootDir = "./dashboard/assets"

func StatusHandler(c http.ResponseWriter, req *http.Request) {
	viewName := "all"

	if val, ok := req.URL.Query()["view"]; ok {
		viewName = val[0]
	}

	var buffer bytes.Buffer
	var services = client.GetServices(viewName)
	var errors, ok = 0, 0
	for _, s := range services {
		if s.State == model.StateError {
			errors++
		} else {
			ok++
		}
	}
	if (req.Header.Get("Accept") == "application/json") {
		ret := struct {
			NumOk    int `json:"num_ok"`
			NumError int `json:"num_error"`
			HasError bool `json:"has_error"`
			Good     bool `json:"good"`
		}{
			ok, errors,
			errors > 0, errors == 0,
		}
		var encoded, _ = json.MarshalIndent(ret, "", "  ")

		c.Header().Add("Content-Type", "application/json")
		c.Header().Add("Content-Length", strconv.Itoa(len(encoded) + 1))
		c.Write(encoded)
		io.WriteString(c, "\n")
	} else {
		buffer.WriteString(fmt.Sprintf("num_ok %d\nnum_error %d\n", ok, errors))
		buffer.WriteString(fmt.Sprintf("has_error %t\ngood %t\n", errors > 0, errors == 0))
		body := buffer.String()
		c.Header().Add("Content-Type", "text/plain")
		c.Header().Add("Content-Length", strconv.Itoa(len(body)))
		io.WriteString(c, body)
	}
}

func DashboardHandler(c http.ResponseWriter, req *http.Request) {
	bytes, err := dashboardHtmlBytes()
	if err == nil {
		c.Header().Set("Content-Type", "text/html")
		c.Write(bytes)
	}
}

func Register(rtr *mux.Router, client_ service.ServiceIf) {
	client = client_
	rtr.HandleFunc("/", DashboardHandler).Methods("GET")
	rtr.HandleFunc("/status", StatusHandler).Methods("GET")
	rtr.PathPrefix("/").Handler(http.FileServer(
		&assetfs.AssetFS{Asset: Asset, AssetDir: AssetDir, AssetInfo: AssetInfo, Prefix: "/"}))
}
