package kibana2

import (
	"context"
	"net/http"

	kbapi "github.com/elastic/terraform-provider-elasticstack/generated/kibana"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// GetDataViews reads all data views from the API.
func GetDataViews(ctx context.Context, client *Client, spaceID string) ([]kbapi.GetDataViewsResponseItem, diag.Diagnostics) {
	resp, err := client.API.GetDataViewsWithResponse(ctx, spaceID)
	if err != nil {
		return nil, utils.FrameworkDiagFromError(err)
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
	resp, err := client.API.GetDataViewWithResponse(ctx, spaceID, viewID)
	if err != nil {
		return nil, utils.FrameworkDiagFromError(err)
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
func CreateDataView(ctx context.Context, client *Client, spaceID string, req kbapi.CreateDataViewJSONRequestBody) (*kbapi.DataViewsDataViewResponseObject, diag.Diagnostics) {
	resp, err := client.API.CreateDataViewWithResponse(ctx, spaceID, req)
	if err != nil {
		return nil, utils.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return resp.JSON200, nil
	default:
		return nil, reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// UpdateDataView updates an existing data view.
func UpdateDataView(ctx context.Context, client *Client, spaceID string, viewID string, req kbapi.UpdateDataViewJSONRequestBody) (*kbapi.DataViewsDataViewResponseObject, diag.Diagnostics) {
	resp, err := client.API.UpdateDataViewWithResponse(ctx, spaceID, viewID, req)
	if err != nil {
		return nil, utils.FrameworkDiagFromError(err)
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
	resp, err := client.API.DeleteDataViewWithResponse(ctx, spaceID, viewID)
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