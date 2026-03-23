// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package kibanaoapi

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// CreateListIndex creates the .lists and .items data streams for a space if they don't exist.
// This is required before any list operations can be performed.
// Returns true if acknowledged, and diagnostics if there was an error.
func CreateListIndex(ctx context.Context, client *Client, spaceID string) (bool, diag.Diagnostics) {
	resp, err := client.API.CreateListIndexWithResponse(ctx, spaceID)
	if err != nil {
		return false, diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 != nil {
			return resp.JSON200.Acknowledged, nil
		}
		return true, nil
	case http.StatusConflict:
		// Data streams already exist ([docs](https://www.elastic.co/docs/api/doc/kibana/operation/operation-createlistindex#operation-createlistindex-409))
		return true, nil
	default:
		return false, reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// ReadListIndex reads the status of .lists and .items data streams for a space.
// Returns the status of list_index and list_item_index separately, and diagnostics on error.
func ReadListIndex(ctx context.Context, client *Client, spaceID string) (listIndex bool, listItemIndex bool, diags diag.Diagnostics) {
	resp, err := client.API.ReadListIndexWithResponse(ctx, spaceID)
	if err != nil {
		return false, false, diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 != nil {
			return resp.JSON200.ListIndex, resp.JSON200.ListItemIndex, nil
		}
		return false, false, nil
	case http.StatusNotFound:
		// Data streams don't exist
		return false, false, nil
	default:
		return false, false, reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// DeleteListIndex deletes the .lists and .items data streams for a space.
// Returns diagnostics if there was an error.
func DeleteListIndex(ctx context.Context, client *Client, spaceID string) diag.Diagnostics {
	resp, err := client.API.DeleteListIndexWithResponse(ctx, spaceID)
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

// GetList reads a security list from the API by ID
func GetList(ctx context.Context, client *Client, spaceID string, params *kbapi.ReadListParams) (*kbapi.SecurityListsAPIList, diag.Diagnostics) {
	resp, err := client.API.ReadListWithResponse(ctx, spaceID, params)
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return nil, diag.Diagnostics{
				diag.NewErrorDiagnostic("Failed to parse list response", "API returned 200 but JSON200 is nil"),
			}
		}
		return resp.JSON200, nil
	case http.StatusNotFound:
		return nil, nil
	default:
		return nil, reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// CreateList creates a new security list.
func CreateList(ctx context.Context, client *Client, spaceID string, body kbapi.CreateListJSONRequestBody) (*kbapi.SecurityListsAPIList, diag.Diagnostics) {
	resp, err := client.API.CreateListWithResponse(ctx, spaceID, body)
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return nil, diag.Diagnostics{
				diag.NewErrorDiagnostic("Failed to parse list response", "API returned 200 but JSON200 is nil"),
			}
		}
		return resp.JSON200, nil
	default:
		return nil, reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// UpdateList updates an existing security list.
func UpdateList(ctx context.Context, client *Client, spaceID string, body kbapi.UpdateListJSONRequestBody) (*kbapi.SecurityListsAPIList, diag.Diagnostics) {
	resp, err := client.API.UpdateListWithResponse(ctx, spaceID, body)
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return nil, diag.Diagnostics{
				diag.NewErrorDiagnostic("Failed to parse list response", "API returned 200 but JSON200 is nil"),
			}
		}
		return resp.JSON200, nil
	default:
		return nil, reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// DeleteList deletes an existing security list.
func DeleteList(ctx context.Context, client *Client, spaceID string, params *kbapi.DeleteListParams) diag.Diagnostics {
	resp, err := client.API.DeleteListWithResponse(ctx, spaceID, params)
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
// The response can be a single item or an array, so we unmarshal from the body.
// When querying by ID, we expect a single item.
func GetListItem(ctx context.Context, client *Client, spaceID string, params *kbapi.ReadListItemParams) (*kbapi.SecurityListsAPIListItem, diag.Diagnostics) {
	resp, err := client.API.ReadListItemWithResponse(ctx, spaceID, params)
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		var listItem kbapi.SecurityListsAPIListItem
		if err := json.Unmarshal(resp.Body, &listItem); err != nil {
			return nil, diag.Diagnostics{
				diag.NewErrorDiagnostic("Failed to parse list item response", err.Error()),
			}
		}

		return &listItem, nil
	case http.StatusNotFound:
		return nil, nil
	default:
		return nil, reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// CreateListItem creates a new security list item.
func CreateListItem(ctx context.Context, client *Client, spaceID string, body kbapi.CreateListItemJSONRequestBody) (*kbapi.SecurityListsAPIListItem, diag.Diagnostics) {
	resp, err := client.API.CreateListItemWithResponse(ctx, spaceID, body)
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return nil, diag.Diagnostics{
				diag.NewErrorDiagnostic("Failed to parse list item response", "API returned 200 but JSON200 is nil"),
			}
		}
		return resp.JSON200, nil
	default:
		return nil, reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// UpdateListItem updates an existing security list item.
func UpdateListItem(ctx context.Context, client *Client, spaceID string, body kbapi.UpdateListItemJSONRequestBody) (*kbapi.SecurityListsAPIListItem, diag.Diagnostics) {
	resp, err := client.API.UpdateListItemWithResponse(ctx, spaceID, body)
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return nil, diag.Diagnostics{
				diag.NewErrorDiagnostic("Failed to parse list item response", "API returned 200 but JSON200 is nil"),
			}
		}
		return resp.JSON200, nil
	default:
		return nil, reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// DeleteListItem deletes an existing security list item.
func DeleteListItem(ctx context.Context, client *Client, spaceID string, params *kbapi.DeleteListItemParams) diag.Diagnostics {
	resp, err := client.API.DeleteListItemWithResponse(ctx, spaceID, params)
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
