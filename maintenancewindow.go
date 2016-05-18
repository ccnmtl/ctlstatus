package ctlstatus

import "time"

type MaintenanceWindow struct {
	Key         string
	Summary     string
	Start       time.Time
	End         time.Time
	Description string
}
