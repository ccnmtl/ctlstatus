package ctlstatus

import "time"

type MaintenanceWindow struct {
	Id          int64
	Summary     string
	Start       time.Time
	End         time.Time
	Description string
}
