package dashboard

import (
	"bytes"
	"github.com/gorilla/mux"
	"net/http"
)

var version string

func DashboardHandler(c http.ResponseWriter, req *http.Request) {
	asset, err := appHtml()
	if err == nil {
		c.Header().Set("Content-Type", "text/html")
		c.Write(asset.bytes)
	}
}

func BundleHandler(c http.ResponseWriter, req *http.Request) {
	asset, err := bundleJs()
	if err == nil {
		c.Header().Set("Content-Type", "text/html")
		c.Write(bytes.Replace(asset.bytes, []byte("%LOVEBEAT_VERSION%"), []byte(version), 1))
	}
}

func Register(rtr *mux.Router, version_ string) {
	version = version_
	rtr.HandleFunc("/", DashboardHandler).Methods("GET")
	rtr.HandleFunc("/views/{name}", DashboardHandler).Methods("GET")
	rtr.HandleFunc("/bundle.js", BundleHandler).Methods("GET")
}
