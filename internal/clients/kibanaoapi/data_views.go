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
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanautil"
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
		return nil, diagutil.ReportUnknownHTTPError(resp.StatusCode(), resp.Body)
	}
}

// GetDataView reads a specific data view from the API.
func GetDataView(ctx context.Context, client *Client, spaceID string, viewID string) (*kbapi.DataViewsDataViewResponseObject, diag.Diagnostics) {
	resp, err := client.API.GetDataViewDefaultWithResponse(ctx, spaceID, viewID)
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	return handleGetTypedResponse(resp.StatusCode(), resp.Body,
		func() *kbapi.DataViewsDataViewResponseObject { return resp.JSON200 })
}

// CreateDataView creates a new data view.
func CreateDataView(ctx context.Context, client *Client, spaceID string, req kbapi.DataViewsCreateDataViewRequestObject) (*kbapi.DataViewsDataViewResponseObject, diag.Diagnostics) {
	resp, err := client.API.CreateDataViewDefaultwWithResponse(ctx, spaceID, req)
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	return handleMutateTypedResponse(resp.StatusCode(), resp.Body,
		func() *kbapi.DataViewsDataViewResponseObject { return resp.JSON200 })
}

// UpdateDataView updates an existing data view.
func UpdateDataView(ctx context.Context, client *Client, spaceID string, viewID string, req kbapi.DataViewsUpdateDataViewRequestObject) (*kbapi.DataViewsDataViewResponseObject, diag.Diagnostics) {
	resp, err := client.API.UpdateDataViewDefaultWithResponse(ctx, spaceID, viewID, req)
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	return handleMutateTypedResponse(resp.StatusCode(), resp.Body,
		func() *kbapi.DataViewsDataViewResponseObject { return resp.JSON200 })
}

// UpdateFieldMetadata writes field metadata (e.g. custom_label) for a data view via
// POST /api/data_views/data_view/{viewId}/fields. It applies space-aware path routing.
func UpdateFieldMetadata(ctx context.Context, client *Client, spaceID string, viewID string, fields map[string]any) diag.Diagnostics {
	resp, err := client.API.UpdateFieldsMetadataDefaultWithResponse(
		ctx,
		viewID,
		kbapi.UpdateFieldsMetadataDefaultJSONRequestBody{Fields: fields},
		kibanautil.SpaceAwarePathRequestEditor(spaceID),
	)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	return diagutil.HandleStatusResponse(resp.StatusCode(), resp.Body, http.StatusOK)
}

// DeleteDataView deletes an existing data view.
func DeleteDataView(ctx context.Context, client *Client, spaceID string, viewID string) diag.Diagnostics {
	resp, err := client.API.DeleteDataViewDefaultWithResponse(ctx, spaceID, viewID)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	return diagutil.HandleStatusResponse(resp.StatusCode(), resp.Body, http.StatusOK, http.StatusNotFound)
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
		return nil, diagutil.ReportUnknownHTTPError(resp.StatusCode(), resp.Body)
	}
}

// SetDefaultDataView sets the default data view.
func SetDefaultDataView(ctx context.Context, client *Client, spaceID string, req kbapi.SetDefaultDatailViewDefaultJSONRequestBody) diag.Diagnostics {
	resp, err := client.API.SetDefaultDatailViewDefaultWithResponse(ctx, spaceID, req)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	return diagutil.HandleStatusResponse(resp.StatusCode(), resp.Body, http.StatusOK)
}
