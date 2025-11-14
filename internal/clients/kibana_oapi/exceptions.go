package kibana_oapi

import (
	"context"
	"net/http"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// GetExceptionList reads an exception list from the API by ID or list_id
func GetExceptionList(ctx context.Context, client *Client, params *kbapi.ReadExceptionListParams) (*kbapi.ReadExceptionListResponse, diag.Diagnostics) {
	resp, err := client.API.ReadExceptionListWithResponse(ctx, params)
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return resp, nil
	case http.StatusNotFound:
		return nil, nil
	default:
		return nil, reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// CreateExceptionList creates a new exception list.
func CreateExceptionList(ctx context.Context, client *Client, body kbapi.CreateExceptionListJSONRequestBody) (*kbapi.CreateExceptionListResponse, diag.Diagnostics) {
	resp, err := client.API.CreateExceptionListWithResponse(ctx, body)
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return resp, nil
	default:
		return nil, reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// UpdateExceptionList updates an existing exception list.
func UpdateExceptionList(ctx context.Context, client *Client, body kbapi.UpdateExceptionListJSONRequestBody) (*kbapi.UpdateExceptionListResponse, diag.Diagnostics) {
	resp, err := client.API.UpdateExceptionListWithResponse(ctx, body)
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return resp, nil
	default:
		return nil, reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// DeleteExceptionList deletes an existing exception list.
func DeleteExceptionList(ctx context.Context, client *Client, params *kbapi.DeleteExceptionListParams) diag.Diagnostics {
	resp, err := client.API.DeleteExceptionListWithResponse(ctx, params)
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

// GetExceptionListItem reads an exception list item from the API by ID or item_id
func GetExceptionListItem(ctx context.Context, client *Client, params *kbapi.ReadExceptionListItemParams) (*kbapi.ReadExceptionListItemResponse, diag.Diagnostics) {
	resp, err := client.API.ReadExceptionListItemWithResponse(ctx, params)
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return resp, nil
	case http.StatusNotFound:
		return nil, nil
	default:
		return nil, reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// CreateExceptionListItem creates a new exception list item.
func CreateExceptionListItem(ctx context.Context, client *Client, body kbapi.CreateExceptionListItemJSONRequestBody) (*kbapi.CreateExceptionListItemResponse, diag.Diagnostics) {
	resp, err := client.API.CreateExceptionListItemWithResponse(ctx, body)
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return resp, nil
	default:
		return nil, reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// UpdateExceptionListItem updates an existing exception list item.
func UpdateExceptionListItem(ctx context.Context, client *Client, body kbapi.UpdateExceptionListItemJSONRequestBody) (*kbapi.UpdateExceptionListItemResponse, diag.Diagnostics) {
	resp, err := client.API.UpdateExceptionListItemWithResponse(ctx, body)
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return resp, nil
	default:
		return nil, reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// DeleteExceptionListItem deletes an existing exception list item.
func DeleteExceptionListItem(ctx context.Context, client *Client, params *kbapi.DeleteExceptionListItemParams) diag.Diagnostics {
	resp, err := client.API.DeleteExceptionListItemWithResponse(ctx, params)
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
