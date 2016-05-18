package ctlstatus

import (
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"text/template"
	"time"

	"appengine"
	"appengine/datastore"
	"appengine/user"
)

type Incident struct {
	Key         string
	Status      string
	Start       time.Time
	End         time.Time
	Summary     string
	Description string
}

func (i Incident) Path() string {
	return "/incident/" + i.Key + "/"
}

func (i Incident) StatusOptions() []string {
	options := map[string][]string{
		"investigating": {"partial", "outage", "resolved"},
		"partial":       {"outage", "resolved"},
		"outage":        {"partial", "resolved"},
		"resolved":      {"investigating", "partial", "outage"},
	}
	return options[i.Status]
}

// if it's resolved, total time between start + end
// otherwise, time from start until now
func (i Incident) Duration() time.Duration {
	if i.Status == "resolved" {
		return i.End.Sub(i.Start)
	}
	return time.Since(i.Start)
}

func (i Incident) DisplayDuration() string {
	d := i.Duration()
	m := int(d.Minutes())
	hours := m / 60.0
	minutes := m % 60
	if hours >= 1 {
		return fmt.Sprintf("%d hours %d minutes", hours, minutes)
	}
	return fmt.Sprintf("%d minutes", minutes)
}

func (i Incident) BootstrapClass() string {
	classes := map[string]string{
		"investigating": "warning",
		"partial":       "danger",
		"outage":        "danger",
		"resolved":      "success",
	}
	return classes[i.Status]
}

func (i Incident) Updates(ctx appengine.Context, k *datastore.Key) ([]Update, error) {
	q := datastore.NewQuery("Update").Ancestor(k).Order("-Timestamp").Limit(100)
	updates := make([]Update, 0, 100)
	_, err := q.GetAll(ctx, &updates)
	if err != nil {
		return []Update{}, err
	}
	return updates, nil
}

type Update struct {
	Incident  *datastore.Key
	Status    string
	Timestamp time.Time
	Comment   string
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

	start, err := time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", r.FormValue("start"))
	if err != nil {
		start = incident.Start
	}

	end, err := time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", r.FormValue("end"))
	if err != nil {
		end = incident.End
	}

	incident.Status = r.FormValue("status")
	incident.Summary = r.FormValue("summary")
	incident.Description = r.FormValue("description")
	incident.Start = start
	incident.End = end

	_, err = datastore.Put(ctx, k, &incident)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, incident.Path(), http.StatusFound)
}
