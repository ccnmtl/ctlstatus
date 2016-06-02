package ctlstatus

import (
	"fmt"
	"html/template"
	"strconv"
	"time"

	"github.com/russross/blackfriday"
)

type MaintenanceWindow struct {
	Id          int64
	Summary     string
	Start       time.Time
	End         time.Time
	Description string
}

func (i MaintenanceWindow) Path() string {
	return "/maintenance/" + strconv.FormatInt(i.Id, 10) + "/"
}

// if it's resolved, total time between start + end
// otherwise, time from start until now
func (i MaintenanceWindow) Duration() time.Duration {
	return i.End.Sub(i.Start)
}

func (i MaintenanceWindow) DisplayDuration() string {
	d := i.Duration()
	m := int(d.Minutes())
	hours := m / 60.0
	minutes := m % 60
	if hours >= 1 {
		return fmt.Sprintf("%d hours %d minutes", hours, minutes)
	}
	return fmt.Sprintf("%d minutes", minutes)
}

func (i MaintenanceWindow) Status() string {
	now := time.Now()
	if now.Before(i.Start) {
		return "upcoming"
	}
	if now.After(i.End) {
		return "completed"
	}
	return "ongoing"
}

func (i MaintenanceWindow) BootstrapClass() string {
	return BootstrapClassFromStatus(i.Status())
}

func (i MaintenanceWindow) StartDate() string {
	return i.Start.In(NYC).Format("Jan 2")
}

func (i MaintenanceWindow) DisplayStart() string {
	return i.Start.In(NYC).Format("Mon Jan 2 15:04")
}

func (i MaintenanceWindow) DisplayEnd() string {
	return i.End.In(NYC).Format("Mon Jan 2 15:04")
}

func (i MaintenanceWindow) EditStart() string {
	return i.Start.In(NYC).Format("2006-01-02 15:04 -0700 MST")
}

func (i MaintenanceWindow) EditEnd() string {
	return i.End.In(NYC).Format("2006-01-02 15:04 -0700 MST")
}

func (i MaintenanceWindow) RenderDescription() template.HTML {
	return template.HTML(string(blackfriday.MarkdownCommon([]byte(i.Description))))
}
