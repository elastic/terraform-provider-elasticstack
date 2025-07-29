package maintenance_window

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type MaintenanceWindowModel struct {
	ID             types.String              `tfsdk:"id"`
	SpaceID        types.String              `tfsdk:"space_id"`
	Title          types.String              `tfsdk:"title"`
	Enabled        types.Bool                `tfsdk:"enabled"`
	CustomSchedule MaintenanceWindowSchedule `tfsdk:"custom_schedule"`
	Scope          MaintenanceWindowScope    `tfsdk:"scope"`
}

type MaintenanceWindowScope struct {
	Alerting MaintenanceWindowAlertingScope `tfsdk:"alerting"`
}

type MaintenanceWindowAlertingScope struct {
	Kql types.String `tfsdk:"kql"`
}

type MaintenanceWindowSchedule struct {
	Start    types.String `tfsdk:"start"`
	Duration types.String `tfsdk:"duration"`
	Timezone types.String `tfsdk:"timezone"`
	// Recurring *MaintenanceWindowScheduleRecurring
}

// type MaintenanceWindowScheduleRecurring struct {
// 	End        *string
// 	Every      *string
// 	OnWeekDay  *[]string
// 	OnMonthDay *[]float32
// 	OnMonth    *[]float32
// }
