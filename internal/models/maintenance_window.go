package models

type MaintenanceWindow struct {
	Id                  string
	MaintenanceWindowID string
	Title               string
	Enabled             bool
	Start               string
	Duration            int
}
