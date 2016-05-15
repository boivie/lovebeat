package dashboard

import (
	assetfs "github.com/elazarl/go-bindata-assetfs"
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

func Register(rtr *mux.Router) {
	rtr.HandleFunc("/", DashboardHandler).Methods("GET")
	rtr.HandleFunc("/views/{name}", DashboardHandler).Methods("GET")
	rtr.PathPrefix("/").Handler(http.FileServer(
		&assetfs.AssetFS{Asset: Asset, AssetDir: AssetDir, AssetInfo: AssetInfo, Prefix: "/"}))
}
