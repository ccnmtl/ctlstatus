package ctlstatus

import (
	"fmt"
	"html/template"
	"strconv"
	"time"

	"github.com/russross/blackfriday"

	"appengine"
	"appengine/datastore"
)

type Incident struct {
	Id          int64
	Status      string
	Start       time.Time
	End         time.Time
	Summary     string
	Description string
}

func (i Incident) Path() string {
	return "/incident/" + strconv.FormatInt(i.Id, 10) + "/"
}

func (i Incident) StatusOptions() []string {
	options := map[string][]string{
		"investigating": {"outage", "resolved"},
		"outage":        {"resolved"},
		"resolved":      {"investigating", "outage"},
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
		"outage":        "danger",
		"resolved":      "success",
		"completed":     "success",
		"ongoing":       "danger",
		"upcoming":      "info",
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
	return i.Start.In(NYC).Format("Jan 2")
}

func (i Incident) DisplayStart() string {
	return i.Start.In(NYC).Format("Mon Jan 2 15:04")
}

func (i Incident) DisplayEnd() string {
	return i.End.In(NYC).Format("Mon Jan 2 15:04")
}

func (i Incident) EditStart() string {
	return i.Start.In(NYC).Format("2006-01-02 15:04 -0700 MST")
}

func (i Incident) EditEnd() string {
	return i.End.In(NYC).Format("2006-01-02 15:04 -0700 MST")
}

func (i Incident) RenderDescription() template.HTML {
	return template.HTML(string(blackfriday.MarkdownCommon([]byte(i.Description))))
}
