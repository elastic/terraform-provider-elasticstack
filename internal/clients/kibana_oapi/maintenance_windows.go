package kibana_oapi

import (
	"context"
	"net/http"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// GetMaintenanceWindow reads a maintenance window from the API by ID
func GetMaintenanceWindow(ctx context.Context, client *Client, spaceID string, maintenanceWindowID string) (*kbapi.GetMaintenanceWindowIdResponse, diag.Diagnostics) {
	resp, err := client.API.GetMaintenanceWindowIdWithResponse(ctx, spaceID, maintenanceWindowID)

	if err != nil {
		return nil, utils.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return resp, nil
	default:
		return nil, reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// CreateMaintenanceWindow creates a new maintenance window.
func CreateMaintenanceWindow(ctx context.Context, client *Client, spaceID string, body kbapi.PostMaintenanceWindowJSONRequestBody) (*kbapi.PostMaintenanceWindowResponse, diag.Diagnostics) {
	resp, err := client.API.PostMaintenanceWindowWithResponse(ctx, spaceID, body)
	if err != nil {
		return nil, utils.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return resp, nil
	default:
		return nil, reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// UpdateMaintenanceWindow updates an existing maintenance window.
func UpdateMaintenanceWindow(ctx context.Context, client *Client, spaceID string, maintenanceWindowID string, req kbapi.PatchMaintenanceWindowIdJSONRequestBody) (*kbapi.PatchMaintenanceWindowIdResponse, diag.Diagnostics) {
	resp, err := client.API.PatchMaintenanceWindowIdWithResponse(ctx, spaceID, maintenanceWindowID, req)
	if err != nil {
		return nil, utils.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return resp, nil
	default:
		return nil, reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// DeleteMaintenanceWindow deletes an existing maintenance window.
func DeleteMaintenanceWindow(ctx context.Context, client *Client, spaceID string, maintenanceWindowID string) diag.Diagnostics {
	resp, err := client.API.DeleteMaintenanceWindowIdWithResponse(ctx, spaceID, maintenanceWindowID)
	if err != nil {
		return utils.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return nil
	case http.StatusNotFound:
		return nil
	default:
		return reportUnknownError(resp.StatusCode(), resp.Body)
	}
}
