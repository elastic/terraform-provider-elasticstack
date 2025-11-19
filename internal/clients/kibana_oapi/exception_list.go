package kibana_oapi

import (
	"context"
	"net/http"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// CreateExceptionList creates a new exception list.
func CreateExceptionList(ctx context.Context, client *Client, req kbapi.CreateExceptionListJSONRequestBody) (*kbapi.SecurityExceptionsAPIExceptionList, diag.Diagnostics) {
	resp, err := client.API.CreateExceptionListWithResponse(ctx, req)
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

// ReadExceptionList reads a specific exception list from the API.
func ReadExceptionList(ctx context.Context, client *Client, params *kbapi.ReadExceptionListParams) (*kbapi.SecurityExceptionsAPIExceptionList, diag.Diagnostics) {
	resp, err := client.API.ReadExceptionListWithResponse(ctx, params)
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

// UpdateExceptionList updates an existing exception list.
func UpdateExceptionList(ctx context.Context, client *Client, req kbapi.UpdateExceptionListJSONRequestBody) (*kbapi.SecurityExceptionsAPIExceptionList, diag.Diagnostics) {
	resp, err := client.API.UpdateExceptionListWithResponse(ctx, req)
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
