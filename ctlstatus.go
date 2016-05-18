package ctlstatus

import (
	"math/rand"
	"net/http"
	"strings"
	"text/template"
	"time"

	"appengine"
	"appengine/datastore"
	"appengine/user"
)

type Update struct {
	Incident  *datastore.Key
	Status    string
	Timestamp time.Time
	Comment   string
}

func (u Update) BootstrapClass() string {
	return BootstrapClassFromStatus(u.Status)
}

func (u Update) DisplayTimestamp() string {
	return u.Timestamp.Format("Mon Jan 2 15:04 2006")
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
	http.HandleFunc("/incident/", showIncident)
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
	tc["current"] = currentIncident(incidents)
	u := user.Current(ctx)
	tc["user"] = u
	if u == nil {
		url, _ := user.LoginURL(ctx, r.URL.String())
		tc["signin_url"] = url
	} else {
		url, _ := user.LogoutURL(ctx, r.URL.String())
		tc["signout_url"] = url
	}
	tc["user"] = u
	if err := indexTemplate.Execute(w, tc); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func currentIncident(incidents []Incident) *Incident {
	now := time.Now()
	for _, incident := range incidents {
		if incident.End.After(now) {
			return &incident
		}
	}
	return nil
}

func newIncident(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	u := user.Current(ctx)
	if u == nil || !u.Admin {
		http.Error(w, "you must be logged in as an admin", http.StatusForbidden)
		return
	}
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
	update := &Update{
		Incident:  key,
		Status:    incident.Status,
		Timestamp: time.Now(),
		Comment:   "New Incident entered",
	}
	ukey := datastore.NewIncompleteKey(ctx, "Update", key)
	if _, err := datastore.Put(ctx, ukey, update); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, incident.Path(), http.StatusFound)
}

var incidentTemplate = template.Must(template.ParseFiles("templates/base.html",
	"templates/incident.html"))

func showIncident(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	parts := strings.Split(r.URL.String(), "/")
	if len(parts) < 3 {
		http.Error(w, "bad request", 404)
		return
	}
	if len(parts) == 4 && parts[3] == "delete" {
		deleteIncident(w, r)
		return
	}
	if len(parts) == 4 && parts[3] == "update" {
		updateIncident(w, r)
		return
	}

	ikey := parts[2]
	k := datastore.NewKey(ctx, "Incident", ikey, 0, nil)
	var incident Incident
	if err := datastore.Get(ctx, k, &incident); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	tc := make(map[string]interface{})
	tc["incident"] = incident

	updates, err := incident.Updates(ctx, k)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tc["updates"] = updates

	u := user.Current(ctx)
	if u == nil {
		url, _ := user.LoginURL(ctx, r.URL.String())
		tc["signin_url"] = url
	} else {
		url, _ := user.LogoutURL(ctx, r.URL.String())
		tc["signout_url"] = url
	}
	tc["user"] = u
	if err := incidentTemplate.Execute(w, tc); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func deleteIncident(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	u := user.Current(ctx)
	if u == nil || !u.Admin {
		http.Error(w, "you must be logged in as an admin", http.StatusForbidden)
		return
	}
	parts := strings.Split(r.URL.String(), "/")
	if len(parts) < 3 {
		http.Error(w, "bad request", 404)
		return
	}
	ikey := parts[2]
	k := datastore.NewKey(ctx, "Incident", ikey, 0, nil)
	err := datastore.Delete(ctx, k)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

func updateIncident(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	u := user.Current(ctx)
	if u == nil || !u.Admin {
		http.Error(w, "you must be logged in as an admin", http.StatusForbidden)
		return
	}
	parts := strings.Split(r.URL.String(), "/")
	if len(parts) < 3 {
		http.Error(w, "bad request", 404)
		return
	}
	ikey := parts[2]
	k := datastore.NewKey(ctx, "Incident", ikey, 0, nil)
	var incident Incident
	if err := datastore.Get(ctx, k, &incident); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	original_status := incident.Status
	incident.Status = r.FormValue("status")
	incident.Summary = r.FormValue("summary")
	incident.Description = r.FormValue("description")

	start, err := time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", r.FormValue("start"))
	if err != nil {
		start = incident.Start
	}
	incident.Start = start

	end, err := time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", r.FormValue("end"))
	if err != nil {
		end = incident.End
		if incident.Status == "resolved" && original_status != "resolved" {
			end = time.Now()
		}
	}

	incident.End = end

	_, err = datastore.Put(ctx, k, &incident)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	update := &Update{
		Incident:  k,
		Status:    incident.Status,
		Timestamp: time.Now(),
		Comment:   r.FormValue("update"),
	}
	ukey := datastore.NewIncompleteKey(ctx, "Update", k)
	if _, err := datastore.Put(ctx, ukey, update); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, incident.Path(), http.StatusFound)
}
