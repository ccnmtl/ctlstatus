package ctlstatus

import (
	"net/http"
	"text/template"
	"time"

	"google.golang.org/appengine/datastore"
)

type Incident struct {
	Status      string
	Start       time.Time
	End         time.Time
	Summary     string
	Description string
}

type Update struct {
	Incident  *datastore.Key
	Status    string
	timestamp time.Time
	comment   string
}

type MaintenanceWindow struct {
	Summary     string
	Start       time.Time
	End         time.Time
	Description string
}

func init() {
	http.HandleFunc("/", indexHandler)
	//	http.HandleFunc("/incident/new", newIncidentHandler)
	//	http.HandleFunc("/incident/", incidentDetailHandler)
	//	http.HandleFunc("/maintenance/new", newIncidentHandler)
}

var indexTemplate = template.Must(template.ParseFiles("templates/base.html",
	"templates/index.html"))

func indexHandler(w http.ResponseWriter, r *http.Request) {
	tc := make(map[string]interface{})
	if err := indexTemplate.Execute(w, tc); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
