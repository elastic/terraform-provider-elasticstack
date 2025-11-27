package kibana_oapi

import (
	"context"
	"net/http"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// CreateListIndex creates the .lists and .items data streams for a space if they don't exist.
// This is required before any list operations can be performed.
func CreateListIndex(ctx context.Context, client *Client, spaceId string) diag.Diagnostics {
	resp, err := client.API.CreateListIndexWithResponse(ctx, kbapi.SpaceId(spaceId))
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

// GetList reads a security list from the API by ID
func GetList(ctx context.Context, client *Client, spaceId string, params *kbapi.ReadListParams) (*kbapi.ReadListResponse, diag.Diagnostics) {
	resp, err := client.API.ReadListWithResponse(ctx, kbapi.SpaceId(spaceId), params)
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

// CreateList creates a new security list.
func CreateList(ctx context.Context, client *Client, spaceId string, body kbapi.CreateListJSONRequestBody) (*kbapi.CreateListResponse, diag.Diagnostics) {
	resp, err := client.API.CreateListWithResponse(ctx, kbapi.SpaceId(spaceId), body)
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

// UpdateList updates an existing security list.
func UpdateList(ctx context.Context, client *Client, spaceId string, body kbapi.UpdateListJSONRequestBody) (*kbapi.UpdateListResponse, diag.Diagnostics) {
	resp, err := client.API.UpdateListWithResponse(ctx, kbapi.SpaceId(spaceId), body)
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

// DeleteList deletes an existing security list.
func DeleteList(ctx context.Context, client *Client, spaceId string, params *kbapi.DeleteListParams) diag.Diagnostics {
	resp, err := client.API.DeleteListWithResponse(ctx, kbapi.SpaceId(spaceId), params)
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

// GetListItem reads a security list item from the API by ID or list_id and value
func GetListItem(ctx context.Context, client *Client, spaceId string, params *kbapi.ReadListItemParams) (*kbapi.ReadListItemResponse, diag.Diagnostics) {
	resp, err := client.API.ReadListItemWithResponse(ctx, kbapi.SpaceId(spaceId), params)
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

// CreateListItem creates a new security list item.
func CreateListItem(ctx context.Context, client *Client, spaceId string, body kbapi.CreateListItemJSONRequestBody) (*kbapi.CreateListItemResponse, diag.Diagnostics) {
	resp, err := client.API.CreateListItemWithResponse(ctx, kbapi.SpaceId(spaceId), body)
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

// UpdateListItem updates an existing security list item.
func UpdateListItem(ctx context.Context, client *Client, spaceId string, body kbapi.UpdateListItemJSONRequestBody) (*kbapi.UpdateListItemResponse, diag.Diagnostics) {
	resp, err := client.API.UpdateListItemWithResponse(ctx, kbapi.SpaceId(spaceId), body)
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

// DeleteListItem deletes an existing security list item.
func DeleteListItem(ctx context.Context, client *Client, spaceId string, params *kbapi.DeleteListItemParams) diag.Diagnostics {
	resp, err := client.API.DeleteListItemWithResponse(ctx, kbapi.SpaceId(spaceId), params)
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
