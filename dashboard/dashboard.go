package dashboard

import (
	"github.com/boivie/lovebeat-go/service"
	"github.com/gorilla/mux"
	"html/template"
	"net/http"
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

func Register(rtr *mux.Router, services *service.Services) {
	svcs = services
	rtr.HandleFunc("/", DashboardHandler).Methods("GET")
}
