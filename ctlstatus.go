package ctlstatus

import (
	"math/rand"
	"net/http"
	"text/template"
	"time"

	"appengine"
	"appengine/datastore"
)

type Incident struct {
	Key         string
	Status      string
	Start       time.Time
	End         time.Time
	Summary     string
	Description string
}

type Update struct {
	Key       string
	Incident  *datastore.Key
	Status    string
	timestamp time.Time
	comment   string
}

type MaintenanceWindow struct {
	Key         string
	Summary     string
	Start       time.Time
	End         time.Time
	Description string
}

const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890abcdefghijklmnopqrstuvwxyz"

func newKey() string {
	var N = 10
	r := make([]byte, N)
	var i = 0
	for i = 0; i < N; i++ {
		r[i] = chars[rand.Intn(len(chars))]
	}
	return string(r)
}

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
	http.HandleFunc("/", index)
	http.HandleFunc("/incident/new", newIncident)
	//	http.HandleFunc("/incident/", incidentDetailHandler)
	//	http.HandleFunc("/maintenance/new", newMaintenanceWindow)
}

var indexTemplate = template.Must(template.ParseFiles("templates/base.html",
	"templates/index.html"))

func index(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	q := datastore.NewQuery("Incident").Order("-End").Limit(10)
	incidents := make([]Incident, 0, 10)
	_, err := q.GetAll(ctx, &incidents)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tc := make(map[string]interface{})
	tc["incidents"] = incidents
	if err := indexTemplate.Execute(w, tc); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func newIncident(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	k := newKey()
	key := datastore.NewKey(ctx, "Incident", k, 0, nil)
	incident := &Incident{
		Key:         k,
		Status:      "investigating",
		Start:       time.Now(),
		End:         time.Now().Add(time.Duration(24) * time.Hour),
		Summary:     r.FormValue("summary"),
		Description: r.FormValue("description"),
	}

	if _, err := datastore.Put(ctx, key, incident); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusFound)
}
