package models

type MaintenanceWindow struct {
	MaintenanceWindowID string
	Title               string
	Enabled             bool
	Start               string
	Duration            int
}
