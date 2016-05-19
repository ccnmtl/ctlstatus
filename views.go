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

var indexTemplate = template.Must(template.ParseFiles("templates/base.html",
	"templates/index.html"))

func addUserToContext(ctx appengine.Context, tc map[string]interface{}, r *http.Request) map[string]interface{} {
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
	return tc
}

func outageUpdatesInRange(ctx appengine.Context, start time.Time, end time.Time) ([]Update, error) {
	q := datastore.NewQuery("Update").
		Filter("Status =", "outage").
		Filter("Timestamp >", start).
		Filter("Timestamp <", end)
	updates := make([]Update, 0, 1)
	_, err := q.GetAll(ctx, &updates)
	if err != nil {
		return []Update{}, err
	}
	return updates, nil
}

func uniqIncidentKeys(updates []Update) []*datastore.Key {
	seen := make(map[*datastore.Key]bool)
	for _, update := range updates {
		seen[update.Incident] = true
	}
	var keys []*datastore.Key
	for k := range seen {
		keys = append(keys, k)
	}
	return keys
}

func uniqIncidents(ctx appengine.Context, updates []Update) ([]Incident, error) {
	ikeys := uniqIncidentKeys(updates)
	incidents := make([]Incident, len(ikeys), len(ikeys))
	if err := datastore.GetMulti(ctx, ikeys, incidents); err != nil {
		return []Incident{}, err
	}
	return incidents, nil
}

func outageIncidentsInRange(ctx appengine.Context, start time.Time, end time.Time) ([]Incident, error) {
	updates, err := outageUpdatesInRange(ctx, start, end)
	if err != nil {
		return []Incident{}, err
	}
	incidents, err := uniqIncidents(ctx, updates)
	if err != nil {
		return []Incident{}, err
	}
	return incidents, nil
}

func sumDurations(incidents []Incident) float64 {
	sum := 0.0
	for _, incident := range incidents {
		sum += incident.Duration().Minutes()
	}
	return sum
}

func yearlyAvailability(ctx appengine.Context) (float64, error) {
	now := time.Now()
	year_ago := now.Add(-1 * time.Duration(365*24) * time.Hour)
	outage_incidents, err := outageIncidentsInRange(ctx, year_ago, now)
	if err != nil {
		return -1.0, err
	}
	// do the calculation in minutes
	total := 365.0 * 24.0 * 60.0
	sum := sumDurations(outage_incidents)
	return 100.0 * (total - sum) / total, nil
}

func monthlyAvailability(ctx appengine.Context) (float64, error) {
	now := time.Now()
	month_ago := now.Add(-1 * time.Duration(30*24) * time.Hour)
	outage_incidents, err := outageIncidentsInRange(ctx, month_ago, now)
	if err != nil {
		return -1.0, err
	}
	// do the calculation in minutes
	total := 30.0 * 24.0 * 60.0
	sum := sumDurations(outage_incidents)
	return 100.0 * (total - sum) / total, nil
}

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
	yearly_availability, err := yearlyAvailability(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	tc["yearly_availability"] = yearly_availability
	monthly_availability, err := monthlyAvailability(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	tc["monthly_availability"] = monthly_availability
	tc = addUserToContext(ctx, tc, r)
	if err := indexTemplate.Execute(w, tc); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func currentIncident(incidents []Incident) *Incident {
	// assumes that incidents come in sorted most recent first
	now := time.Now()
	for _, incident := range incidents {
		if incident.End.After(now) || incident.Status == "outage" {
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
	if r.Method != "POST" {
		http.Error(w, "bad request", 405)
		return
	}
	summary := r.FormValue("summary")
	if summary == "" {
		http.Error(w, "bad request. need summary", 400)
		return
	}
	k := newKey()
	key := datastore.NewKey(ctx, "Incident", k, 0, nil)
	incident := &Incident{
		Key:         k,
		Status:      r.FormValue("status"),
		Start:       time.Now(),
		End:         time.Now().Add(time.Duration(24) * time.Hour),
		Summary:     summary,
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
	tc = addUserToContext(ctx, tc, r)
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
	if r.Method != "POST" {
		http.Error(w, "bad request", 405)
		return
	}
	parts := strings.Split(r.URL.String(), "/")
	if len(parts) < 3 {
		http.Error(w, "bad request", 404)
		return
	}
	ikey := parts[2]
	k := datastore.NewKey(ctx, "Incident", ikey, 0, nil)

	q := datastore.NewQuery("Update").Ancestor(k).Order("Timestamp").Limit(100)
	updates := make([]Update, 0, 100)
	ukeys, err := q.GetAll(ctx, &updates)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = datastore.DeleteMulti(ctx, ukeys)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = datastore.Delete(ctx, k)
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
	if r.Method != "POST" {
		http.Error(w, "bad request", 405)
		return
	}
	parts := strings.Split(r.URL.String(), "/")
	if len(parts) < 3 {
		http.Error(w, "bad request", 404)
		return
	}
	summary := r.FormValue("summary")
	if summary == "" {
		http.Error(w, "bad request. need summary", 400)
		return
	}
	ikey := parts[2]
	k := datastore.NewKey(ctx, "Incident", ikey, 0, nil)
	var incident Incident
	err := datastore.RunInTransaction(ctx, func(ctx appengine.Context) error {
		if err := datastore.Get(ctx, k, &incident); err != nil {
			return err
		}
		original_status := incident.Status
		incident.Status = r.FormValue("status")
		incident.Summary = summary
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
			return err
		}
		return nil
	}, nil)
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
