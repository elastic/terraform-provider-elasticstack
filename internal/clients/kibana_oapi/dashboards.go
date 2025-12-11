package kibana_oapi

import (
	"context"
	"fmt"
	"net/http"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// buildSpaceAwarePath constructs an API path with space awareness.
// If spaceID is empty or "default", returns the basePath unchanged.
// Otherwise, prepends "/s/{spaceID}" to the basePath.
func buildSpaceAwarePath(spaceID, basePath string) string {
	if spaceID != "" && spaceID != "default" {
		return fmt.Sprintf("/s/%s%s", spaceID, basePath)
	}
	return basePath
}

// spaceAwarePathRequestEditor returns a RequestEditorFn that modifies the request path for space awareness.
func spaceAwarePathRequestEditor(spaceID string) func(ctx context.Context, req *http.Request) error {
	return func(ctx context.Context, req *http.Request) error {
		req.URL.Path = buildSpaceAwarePath(spaceID, req.URL.Path)
		return nil
	}
}

// These headers and query parameters appear to be required by the Dashboard API at the moment.
func addApiVersionQueryParamRequestEditor() func(ctx context.Context, req *http.Request) error {
	return func(ctx context.Context, req *http.Request) error {
		req.Header.Add("x-elastic-internal-origin", "Kibana")
		query := req.URL.Query()
		query.Add("apiVersion", "1")
		query.Add("allowUnmappedKeys", "true")
		req.URL.RawQuery = query.Encode()
		return nil
	}
}

// GetDashboard reads a specific dashboard from the API.
func GetDashboard(ctx context.Context, client *Client, spaceID string, dashboardID string) (*kbapi.GetDashboardsDashboardIdResponse, diag.Diagnostics) {
	resp, err := client.API.GetDashboardsDashboardIdWithResponse(
		ctx, dashboardID,
		spaceAwarePathRequestEditor(spaceID),
		addApiVersionQueryParamRequestEditor(),
	)
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

// CreateDashboard creates a new dashboard.
func CreateDashboard(ctx context.Context, client *Client, spaceID string, req kbapi.PostDashboardsDashboardJSONRequestBody) (*kbapi.PostDashboardsDashboardResponse, diag.Diagnostics) {
	resp, err := client.API.PostDashboardsDashboardWithResponse(
		ctx,
		req,
		spaceAwarePathRequestEditor(spaceID),
		addApiVersionQueryParamRequestEditor(),
	)
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

// UpdateDashboard updates an existing dashboard.
func UpdateDashboard(ctx context.Context, client *Client, spaceID string, dashboardID string, req kbapi.PutDashboardsDashboardIdJSONRequestBody) (*kbapi.PutDashboardsDashboardIdResponse, diag.Diagnostics) {
	resp, err := client.API.PutDashboardsDashboardIdWithResponse(
		ctx, dashboardID, req,
		spaceAwarePathRequestEditor(spaceID),
		addApiVersionQueryParamRequestEditor(),
	)
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

// DeleteDashboard deletes an existing dashboard.
func DeleteDashboard(ctx context.Context, client *Client, spaceID string, dashboardID string) diag.Diagnostics {
	resp, err := client.API.DeleteDashboardsDashboardIdWithResponse(
		ctx, dashboardID,
		spaceAwarePathRequestEditor(spaceID),
		addApiVersionQueryParamRequestEditor(),
	)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK, http.StatusNoContent:
		return nil
	case http.StatusNotFound:
		return nil
	default:
		return reportUnknownError(resp.StatusCode(), resp.Body)
	}
}
