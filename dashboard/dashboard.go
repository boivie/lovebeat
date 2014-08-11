package dashboard

import (
	"github.com/boivie/lovebeat-go/service"
	"github.com/gorilla/mux"
	"html/template"
	"net/http"
)

func DashboardHandler(w http.ResponseWriter, r *http.Request) {
	var services = service.GetServices()

	tc := make(map[string]interface{})
	tc["services"] = services

	templates := template.Must(template.ParseFiles("templates/base.html", "templates/index.html"))
	if err := templates.Execute(w, tc); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func Register(rtr *mux.Router) {
	rtr.HandleFunc("/", DashboardHandler).Methods("GET")
}
