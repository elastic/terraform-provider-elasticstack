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

	return &models.MaintenanceWindow{
		Id:       res.Id,
		SpaceId:  spaceId,
		Title:    res.Title,
		Enabled:  res.Enabled,
		Start:    res.Start,
		Duration: int(res.Duration),
	}
}

func CreateMaintenanceWindow(ctx context.Context, apiClient ApiClient, maintenanceWindow models.MaintenanceWindow) (*models.MaintenanceWindow, diag.Diagnostics) {
	client, err := apiClient.GetAlertingClient()
	if err != nil {
		return nil, diag.FromErr(err)
	}

	ctxWithAuth := apiClient.SetAlertingAuthContext(ctx)

	reqModel := alerting.CreateMaintenanceWindowRequest{
		Title:    maintenanceWindow.Title,
		Enabled:  &maintenanceWindow.Enabled,
		Start:    maintenanceWindow.Start,
		Duration: float32(maintenanceWindow.Duration),
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
			Summary:  "Create rule returned an empty response",
			Detail:   fmt.Sprintf("Create rule returned an empty response with HTTP status code [%d].", res.StatusCode),
		}}
	}

	maintenanceWindow.Id = maintenanceWindowRes.Id

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
