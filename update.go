package ctlstatus

import (
	"html/template"
	"time"

	"github.com/russross/blackfriday"

	"appengine/datastore"
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
	return u.Timestamp.In(NYC).Format("Mon Jan 2 15:04")
}

func (u Update) RenderComment() template.HTML {
	return template.HTML(string(blackfriday.MarkdownCommon([]byte(u.Comment))))
}
