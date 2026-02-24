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
	"net/http"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// GetExceptionList reads an exception list from the API by ID or list_id
func GetExceptionList(ctx context.Context, client *Client, spaceID string, params *kbapi.ReadExceptionListParams) (*kbapi.SecurityExceptionsAPIExceptionList, diag.Diagnostics) {
	resp, err := client.API.ReadExceptionListWithResponse(ctx, spaceID, params)
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

// CreateExceptionList creates a new exception list.
func CreateExceptionList(ctx context.Context, client *Client, spaceID string, body kbapi.CreateExceptionListJSONRequestBody) (*kbapi.SecurityExceptionsAPIExceptionList, diag.Diagnostics) {
	resp, err := client.API.CreateExceptionListWithResponse(ctx, spaceID, body)
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

// UpdateExceptionList updates an existing exception list.
func UpdateExceptionList(ctx context.Context, client *Client, spaceID string, body kbapi.UpdateExceptionListJSONRequestBody) (*kbapi.SecurityExceptionsAPIExceptionList, diag.Diagnostics) {
	resp, err := client.API.UpdateExceptionListWithResponse(ctx, spaceID, body)
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
func DeleteExceptionList(ctx context.Context, client *Client, spaceID string, params *kbapi.DeleteExceptionListParams) diag.Diagnostics {
	resp, err := client.API.DeleteExceptionListWithResponse(ctx, spaceID, params)
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
func GetExceptionListItem(ctx context.Context, client *Client, spaceID string, params *kbapi.ReadExceptionListItemParams) (*kbapi.SecurityExceptionsAPIExceptionListItem, diag.Diagnostics) {
	resp, err := client.API.ReadExceptionListItemWithResponse(ctx, spaceID, params)
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

// CreateExceptionListItem creates a new exception list item.
func CreateExceptionListItem(ctx context.Context, client *Client, spaceID string, body kbapi.CreateExceptionListItemJSONRequestBody) (*kbapi.SecurityExceptionsAPIExceptionListItem, diag.Diagnostics) {
	resp, err := client.API.CreateExceptionListItemWithResponse(ctx, spaceID, body)
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

// UpdateExceptionListItem updates an existing exception list item.
func UpdateExceptionListItem(ctx context.Context, client *Client, spaceID string, body kbapi.UpdateExceptionListItemJSONRequestBody) (*kbapi.SecurityExceptionsAPIExceptionListItem, diag.Diagnostics) {
	resp, err := client.API.UpdateExceptionListItemWithResponse(ctx, spaceID, body)
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
func DeleteExceptionListItem(ctx context.Context, client *Client, spaceID string, params *kbapi.DeleteExceptionListItemParams) diag.Diagnostics {
	resp, err := client.API.DeleteExceptionListItemWithResponse(ctx, spaceID, params)
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
