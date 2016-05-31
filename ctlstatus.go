package ctlstatus

import (
	"math/rand"
	"net/http"
	"time"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
	http.HandleFunc("/", index)
	http.HandleFunc("/incident/new", newIncident)
	http.HandleFunc("/incident/", showIncident)
	http.HandleFunc("/maintenance/new", newMaintenanceWindow)
}
