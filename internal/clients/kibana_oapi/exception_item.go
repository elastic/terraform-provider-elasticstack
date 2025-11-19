package kibana_oapi

import (
	"context"
	"net/http"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// CreateExceptionListItem creates a new exception list item.
func CreateExceptionListItem(ctx context.Context, client *Client, req kbapi.CreateExceptionListItemJSONRequestBody) (*kbapi.SecurityExceptionsAPIExceptionListItem, diag.Diagnostics) {
	resp, err := client.API.CreateExceptionListItemWithResponse(ctx, req)
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

// ReadExceptionListItem reads a specific exception list item from the API.
func ReadExceptionListItem(ctx context.Context, client *Client, params *kbapi.ReadExceptionListItemParams) (*kbapi.SecurityExceptionsAPIExceptionListItem, diag.Diagnostics) {
	resp, err := client.API.ReadExceptionListItemWithResponse(ctx, params)
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

// UpdateExceptionListItem updates an existing exception list item.
func UpdateExceptionListItem(ctx context.Context, client *Client, req kbapi.UpdateExceptionListItemJSONRequestBody) (*kbapi.SecurityExceptionsAPIExceptionListItem, diag.Diagnostics) {
	resp, err := client.API.UpdateExceptionListItemWithResponse(ctx, req)
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
