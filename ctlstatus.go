package ctlstatus

import (
	"fmt"
	"net/http"
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

func indexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "status")
}
