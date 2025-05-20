package models

type MaintenanceWindow struct {
	MaintenanceWindowId string
	SpaceId             string
	Title               string
	Enabled             bool
	Scope               *MaintenanceWindowScope
	CustomSchedule      MaintenanceWindowSchedule
}

type MaintenanceWindowScope struct {
	Alerting *MaintenanceWindowAlertingScope
}

type MaintenanceWindowAlertingScope struct {
	Kql string
}

type MaintenanceWindowSchedule struct {
	Start     string
	Duration  string
	Timezone  *string
	Recurring *MaintenanceWindowScheduleRecurring
}

type MaintenanceWindowScheduleRecurring struct {
	End        *string
	Every      *string
	OnWeekDay  *[]string
	OnMonthDay *[]float32
	OnMonth    *[]float32
}
