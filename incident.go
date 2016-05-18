package ctlstatus

import (
	"fmt"
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

func BootstrapClassFromStatus(s string) string {
	classes := map[string]string{
		"investigating": "warning",
		"partial":       "danger",
		"outage":        "danger",
		"resolved":      "success",
	}
	return classes[s]
}

func (i Incident) BootstrapClass() string {
	return BootstrapClassFromStatus(i.Status)
}

func (i Incident) Updates(ctx appengine.Context, k *datastore.Key) ([]Update, error) {
	q := datastore.NewQuery("Update").Ancestor(k).Order("Timestamp").Limit(100)
	updates := make([]Update, 0, 100)
	_, err := q.GetAll(ctx, &updates)
	if err != nil {
		return []Update{}, err
	}
	return updates, nil
}

func (i Incident) StartDate() string {
	return i.Start.Format("Jan 2")
}

func (i Incident) DisplayStart() string {
	return i.Start.Format("Mon Jan 2 15:04")
}

func (i Incident) DisplayEnd() string {
	return i.End.Format("Mon Jan 2 15:04")
}
