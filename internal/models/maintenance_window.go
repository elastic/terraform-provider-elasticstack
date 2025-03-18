package models

type MaintenanceWindow struct {
	Id                  string
	SpaceId             string
	MaintenanceWindowID string
	Title               string
	Enabled             bool
	Start               string
	Duration            int
}
