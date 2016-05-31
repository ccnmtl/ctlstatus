package ctlstatus

import (
	"math/rand"
	"net/http"
	"time"
)

var NYC *time.Location

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
	http.HandleFunc("/", index)
	http.HandleFunc("/incident/new", newIncident)
	http.HandleFunc("/incident/", showIncident)
	http.HandleFunc("/maintenance/new", newMaintenanceWindow)
	http.HandleFunc("/maintenance/", showMaintenanceWindow)
	NYC, _ = time.LoadLocation("America/New_York")
}
