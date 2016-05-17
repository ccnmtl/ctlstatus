package ctlstatus

import (
	"net/http"
	"text/template"
	"time"

	"appengine"
	"appengine/datastore"
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
	http.HandleFunc("/incident/new", newIncidentHandler)
	//	http.HandleFunc("/incident/", incidentDetailHandler)
	//	http.HandleFunc("/maintenance/new", newIncidentHandler)
}

var indexTemplate = template.Must(template.ParseFiles("templates/base.html",
	"templates/index.html"))

func indexHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	q := datastore.NewQuery("Incident").Order("-End").Limit(10)
	incidents := make([]Incident, 0, 10)
	if _, err := q.GetAll(ctx, &incidents); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tc := make(map[string]interface{})
	tc["incidents"] = incidents
	if err := indexTemplate.Execute(w, tc); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func newIncidentHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	incident := &Incident{
		Status:      "investigating",
		Start:       time.Now(),
		End:         time.Now().Add(time.Duration(24) * time.Hour),
		Summary:     r.FormValue("summary"),
		Description: r.FormValue("description"),
	}
	key := datastore.NewIncompleteKey(ctx, "Incident", nil)
	if _, err := datastore.Put(ctx, key, incident); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusFound)
}
