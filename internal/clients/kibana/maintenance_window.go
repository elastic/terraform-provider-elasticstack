package kibana

import (
	"context"
	"fmt"
	"net/http"

	"github.com/elastic/terraform-provider-elasticstack/generated/alerting"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

func maintenanceWindowResponseToModel(spaceId string, res *alerting.MaintenanceWindowResponseProperties) *models.MaintenanceWindow {
	if res == nil {
		return nil
	}

	var recurring *models.MaintenanceWindowScheduleRecurring

	if alerting.IsNil(res.Schedule.Custom.Recurring) {
		recurring = nil
	} else {
		unwrappedRecurring := unwrapOptionalField(res.Schedule.Custom.Recurring)

		recurring = &models.MaintenanceWindowScheduleRecurring{
			End:         unwrappedRecurring.End,
			Every:       unwrappedRecurring.Every,
			Occurrences: unwrappedRecurring.Occurrences,
			OnWeekDay:   &unwrappedRecurring.OnWeekDay,
			OnMonthDay:  &unwrappedRecurring.OnMonthDay,
			OnMonth:     &unwrappedRecurring.OnMonth,
		}
	}

	var scope *models.MaintenanceWindowScope

	if alerting.IsNil(res.Scope) {
		scope = nil
	} else {
		unwrappedScope := unwrapOptionalField(res.Scope)

		scope = &models.MaintenanceWindowScope{
			Alerting: &models.MaintenanceWindowAlertingScope{
				Kql: unwrappedScope.Alerting.Query.Kql,
			},
		}
	}

	return &models.MaintenanceWindow{
		MaintenanceWindowId: res.Id,
		SpaceId:             spaceId,
		Title:               res.Title,
		Enabled:             res.Enabled,
		CustomSchedule: models.MaintenanceWindowSchedule{
			Start:     res.Schedule.Custom.Start,
			Duration:  res.Schedule.Custom.Duration,
			Timezone:  res.Schedule.Custom.Timezone,
			Recurring: recurring,
		},
		Scope: scope,
	}
}

func CreateMaintenanceWindow(ctx context.Context, apiClient ApiClient, maintenanceWindow models.MaintenanceWindow) (*models.MaintenanceWindow, diag.Diagnostics) {
	client, err := apiClient.GetAlertingClient()
	if err != nil {
		return nil, diag.FromErr(err)
	}

	ctxWithAuth := apiClient.SetAlertingAuthContext(ctx)

	schedule := alerting.CreateMaintenanceWindowRequestSchedule{
		Custom: alerting.CreateMaintenanceWindowRequestScheduleCustom{
			Start:    maintenanceWindow.CustomSchedule.Start,
			Duration: maintenanceWindow.CustomSchedule.Duration,
			Timezone: maintenanceWindow.CustomSchedule.Timezone,
		},
	}

	var recurring *alerting.CreateMaintenanceWindowRequestScheduleCustomRecurring

	if alerting.IsNil(maintenanceWindow.CustomSchedule.Recurring) {
		recurring = nil
	} else {
		unwrappedRecurring := unwrapOptionalField(maintenanceWindow.CustomSchedule.Recurring)
		unwrappedOnWeekDay := unwrapOptionalField(unwrappedRecurring.OnWeekDay)
		unwrappedOnMonthDay := unwrapOptionalField(unwrappedRecurring.OnMonthDay)
		unwrappedOnMonth := unwrapOptionalField(unwrappedRecurring.OnMonth)

		recurring = &alerting.CreateMaintenanceWindowRequestScheduleCustomRecurring{
			End:         unwrappedRecurring.End,
			Every:       unwrappedRecurring.Every,
			Occurrences: unwrappedRecurring.Occurrences,
			OnWeekDay:   unwrappedOnWeekDay,
			OnMonthDay:  unwrappedOnMonthDay,
			OnMonth:     unwrappedOnMonth,
		}
	}

	schedule.Custom.Recurring = recurring

	var scope *alerting.CreateMaintenanceWindowRequestScope

	if alerting.IsNil(maintenanceWindow.Scope) {
		scope = nil
	} else {
		scope = &alerting.CreateMaintenanceWindowRequestScope{
			Alerting: alerting.CreateMaintenanceWindowRequestScopeAlerting{
				Query: alerting.CreateMaintenanceWindowRequestScopeAlertingQuery{
					Kql: maintenanceWindow.Scope.Alerting.Kql,
				},
			},
		}
	}

	reqModel := alerting.CreateMaintenanceWindowRequest{
		Title:    maintenanceWindow.Title,
		Enabled:  &maintenanceWindow.Enabled,
		Schedule: schedule,
		Scope:    scope,
	}

	req := client.CreateMaintenanceWindow(ctxWithAuth, maintenanceWindow.SpaceId).KbnXsrf("true").CreateMaintenanceWindowRequest(reqModel)
	maintenanceWindowRes, res, err := req.Execute()
	if err != nil && res == nil {
		return nil, diag.FromErr(err)
	}

	defer res.Body.Close()

	diags := utils.CheckHttpError(res, "Unable to create maintenance window")
	if diags.HasError() {
		return nil, diags
	}

	if maintenanceWindowRes == nil {
		return nil, diag.Diagnostics{diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Create maintenance window returned an empty response",
			Detail:   fmt.Sprintf("Create maintenance window returned an empty response with HTTP status code [%d].", res.StatusCode),
		}}
	}

	return maintenanceWindowResponseToModel(maintenanceWindow.SpaceId, maintenanceWindowRes), nil
}

func GetMaintenanceWindow(ctx context.Context, apiClient *clients.ApiClient, maintenanceWindowId string, spaceId string) (*models.MaintenanceWindow, diag.Diagnostics) {
	client, err := apiClient.GetAlertingClient()
	if err != nil {
		return nil, diag.FromErr(err)
	}

	ctxWithAuth := apiClient.SetAlertingAuthContext(ctx)

	req := client.GetMaintenanceWindow(ctxWithAuth, maintenanceWindowId, spaceId)
	maintenanceWindowRes, res, err := req.Execute()

	if err != nil && res == nil {
		return nil, diag.FromErr(err)
	}

	defer res.Body.Close()
	if res.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	return maintenanceWindowResponseToModel(spaceId, maintenanceWindowRes), utils.CheckHttpError(res, "Unable to get maintenance window")
}

func DeleteMaintenanceWindow(ctx context.Context, apiClient *clients.ApiClient, maintenanceWindowId string, spaceId string) diag.Diagnostics {
	client, err := apiClient.GetAlertingClient()
	if err != nil {
		return diag.FromErr(err)
	}

	ctxWithAuth := apiClient.SetAlertingAuthContext(ctx)

	req := client.DeleteMaintenanceWindow(ctxWithAuth, maintenanceWindowId, spaceId).KbnXsrf("true")
	res, err := req.Execute()
	if err != nil && res == nil {
		return diag.FromErr(err)
	}

	defer res.Body.Close()
	return utils.CheckHttpError(res, "Unable to delete maintenance window")
}

func UpdateMaintenanceWindow(ctx context.Context, apiClient ApiClient, maintenanceWindow models.MaintenanceWindow) (*models.MaintenanceWindow, diag.Diagnostics) {
	client, err := apiClient.GetAlertingClient()
	if err != nil {
		return nil, diag.FromErr(err)
	}

	ctxWithAuth := apiClient.SetAlertingAuthContext(ctx)

	schedule := alerting.CreateMaintenanceWindowRequestSchedule{
		Custom: alerting.CreateMaintenanceWindowRequestScheduleCustom{
			Start:    maintenanceWindow.CustomSchedule.Start,
			Duration: maintenanceWindow.CustomSchedule.Duration,
			Timezone: maintenanceWindow.CustomSchedule.Timezone,
		},
	}

	var recurring *alerting.CreateMaintenanceWindowRequestScheduleCustomRecurring

	if alerting.IsNil(maintenanceWindow.CustomSchedule.Recurring) {
		recurring = nil
	} else {
		unwrappedRecurring := unwrapOptionalField(maintenanceWindow.CustomSchedule.Recurring)
		unwrappedOnWeekDay := unwrapOptionalField(unwrappedRecurring.OnWeekDay)
		unwrappedOnMonthDay := unwrapOptionalField(unwrappedRecurring.OnMonthDay)
		unwrappedOnMonth := unwrapOptionalField(unwrappedRecurring.OnMonth)

		recurring = &alerting.CreateMaintenanceWindowRequestScheduleCustomRecurring{
			End:         unwrappedRecurring.End,
			Every:       unwrappedRecurring.Every,
			Occurrences: unwrappedRecurring.Occurrences,
			OnWeekDay:   unwrappedOnWeekDay,
			OnMonthDay:  unwrappedOnMonthDay,
			OnMonth:     unwrappedOnMonth,
		}
	}

	schedule.Custom.Recurring = recurring

	var scope *alerting.CreateMaintenanceWindowRequestScope

	if alerting.IsNil(maintenanceWindow.Scope) {
		scope = nil
	} else {
		scope = &alerting.CreateMaintenanceWindowRequestScope{
			Alerting: alerting.CreateMaintenanceWindowRequestScopeAlerting{
				Query: alerting.CreateMaintenanceWindowRequestScopeAlertingQuery{
					Kql: maintenanceWindow.Scope.Alerting.Kql,
				},
			},
		}
	}

	reqModel := alerting.UpdateMaintenanceWindowRequest{
		Title:    &maintenanceWindow.Title,
		Enabled:  &maintenanceWindow.Enabled,
		Schedule: &schedule,
		Scope:    scope,
	}

	req := client.UpdateMaintenanceWindow(ctxWithAuth, maintenanceWindow.MaintenanceWindowId, maintenanceWindow.SpaceId).KbnXsrf("true").UpdateMaintenanceWindowRequest(reqModel)

	maintenanceWindowRes, res, err := req.Execute()
	if err != nil && res == nil {
		return nil, diag.FromErr(err)
	}

	defer res.Body.Close()

	if diags := utils.CheckHttpError(res, "Unable to update maintenance window"); diags.HasError() {
		return nil, diags
	}

	if maintenanceWindowRes == nil {
		return nil, diag.Diagnostics{diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Update maintenance window returned an empty response",
			Detail:   fmt.Sprintf("Update maintenance window returned an empty response with HTTP status code [%d].", res.StatusCode),
		}}
	}

	return maintenanceWindowResponseToModel(maintenanceWindow.SpaceId, maintenanceWindowRes), nil
}
