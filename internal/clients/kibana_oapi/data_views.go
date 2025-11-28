package kibana_oapi

import (
	"context"
	"net/http"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// GetDataViews reads all data views from the API.
func GetDataViews(ctx context.Context, client *Client, spaceID string) ([]kbapi.GetDataViewsResponseItem, diag.Diagnostics) {
	resp, err := client.API.GetAllDataViewsDefaultWithResponse(ctx, spaceID)
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return *resp.JSON200.DataView, nil
	default:
		return nil, reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// GetDataView reads a specific data view from the API.
func GetDataView(ctx context.Context, client *Client, spaceID string, viewID string) (*kbapi.DataViewsDataViewResponseObject, diag.Diagnostics) {
	resp, err := client.API.GetDataViewDefaultWithResponse(ctx, spaceID, viewID)
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return resp.JSON200, nil
	case http.StatusNotFound:
		return nil, nil
	default:
		return nil, reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// CreateDataView creates a new data view.
func CreateDataView(ctx context.Context, client *Client, spaceID string, req kbapi.DataViewsCreateDataViewRequestObject) (*kbapi.DataViewsDataViewResponseObject, diag.Diagnostics) {
	resp, err := client.API.CreateDataViewDefaultwWithResponse(ctx, spaceID, req)
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return resp.JSON200, nil
	default:
		return nil, reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// UpdateDataView updates an existing data view.
func UpdateDataView(ctx context.Context, client *Client, spaceID string, viewID string, req kbapi.DataViewsUpdateDataViewRequestObject) (*kbapi.DataViewsDataViewResponseObject, diag.Diagnostics) {
	resp, err := client.API.UpdateDataViewDefaultWithResponse(ctx, spaceID, viewID, req)
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return resp.JSON200, nil
	default:
		return nil, reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// DeleteDataView deletes an existing data view.
func DeleteDataView(ctx context.Context, client *Client, spaceID string, viewID string) diag.Diagnostics {
	resp, err := client.API.DeleteDataViewDefaultWithResponse(ctx, spaceID, viewID)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
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

// GetDefaultDataView reads the default data view from the API.
func GetDefaultDataView(ctx context.Context, client *Client, spaceID string) (*string, diag.Diagnostics) {
	resp, err := client.API.GetDefaultDataViewDefaultWithResponse(ctx, spaceID)
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	// We don't check for a 404 here. The API doesn't document a 404 response for this endpoint.
	// In testing, there's no case where a 404 is returned:
	// - If no default data view is set, the API returns 200 with an empty string as the data view ID.
	// - If the space doesn't exist, the API still returns 200 with an empty string as the data view ID.
	// Therefore, we only handle the 200 response and treat any other status code as an error.
	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 != nil && resp.JSON200.DataViewId != nil && *resp.JSON200.DataViewId != "" {
			return resp.JSON200.DataViewId, nil
		}
		return nil, nil
	default:
		return nil, reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// SetDefaultDataView sets the default data view.
func SetDefaultDataView(ctx context.Context, client *Client, spaceID string, req kbapi.SetDefaultDatailViewDefaultJSONRequestBody) diag.Diagnostics {
	resp, err := client.API.SetDefaultDatailViewDefaultWithResponse(ctx, spaceID, req)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return nil
	default:
		return reportUnknownError(resp.StatusCode(), resp.Body)
	}
}
