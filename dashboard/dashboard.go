package dashboard

import (
	"github.com/gorilla/mux"
	"net/http"
)

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
		c.Write(asset.bytes)
	}
}

func Register(rtr *mux.Router) {
	rtr.HandleFunc("/", DashboardHandler).Methods("GET")
	rtr.HandleFunc("/views/{name}", DashboardHandler).Methods("GET")
	rtr.HandleFunc("/bundle.js", BundleHandler).Methods("GET")
}
