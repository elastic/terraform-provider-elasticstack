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
	"fmt"
	"net/http"
	"slices"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanautil"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

type updateObjectsSpacesResponse struct {
	Objects []updateObjectsSpacesObject `json:"objects"`
}

type updateObjectsSpacesObject struct {
	ID    string                          `json:"id"`
	Type  string                          `json:"type"`
	Error *updateObjectsSpacesObjectError `json:"error"`
}

type updateObjectsSpacesObjectError struct {
	StatusCode int    `json:"statusCode"`
	Error      string `json:"error"`
	Message    string `json:"message"`
}

// UpdateDataViewNamespaces diffs old/new namespaces and calls
func UpdateDataViewNamespaces(
	ctx context.Context,
	client *Client,
	spaceID string,
	dataViewID string,
	oldNamespaces []string,
	newNamespaces []string,
) diag.Diagnostics {
	var diags diag.Diagnostics

	spacesToAdd := []string{}
	spacesToRemove := []string{}

	for _, ns := range newNamespaces {
		if !slices.Contains(oldNamespaces, ns) {
			spacesToAdd = append(spacesToAdd, ns)
		}
	}
	for _, ns := range oldNamespaces {
		if !slices.Contains(newNamespaces, ns) {
			spacesToRemove = append(spacesToRemove, ns)
		}
	}

	if len(spacesToAdd) == 0 && len(spacesToRemove) == 0 {
		return diags
	}

	reqBody := kbapi.PostSpacesUpdateObjectsSpacesJSONRequestBody{
		Objects: []struct {
			Id   string `json:"id"` //nolint:revive // var-naming: generated API struct
			Type string `json:"type"`
		}{
			{Id: dataViewID, Type: "index-pattern"},
		},
		SpacesToAdd:    spacesToAdd,
		SpacesToRemove: spacesToRemove,
	}

	resp, err := client.API.PostSpacesUpdateObjectsSpacesWithResponse(
		ctx,
		reqBody,
		kibanautil.SpaceAwarePathRequestEditor(spaceID),
	)
	if err != nil {
		diags.AddError("Failed to update data view namespaces", err.Error())
		return diags
	}

	if resp.StatusCode() != http.StatusOK {
		diags.AddError(
			"Unexpected status when updating data view namespaces",
			string(resp.Body),
		)
		return diags
	}

	var parsed updateObjectsSpacesResponse
	if err := json.Unmarshal(resp.Body, &parsed); err != nil {
		diags.AddError(
			"Unexpected response when updating data view namespaces",
			fmt.Sprintf("failed to parse response body: %s", err.Error()),
		)
		return diags
	}

	if len(parsed.Objects) == 0 {
		diags.AddError(
			"Unexpected response from data view namespace update",
			responseBodySnippet(resp.Body),
		)
		return diags
	}

	for _, obj := range parsed.Objects {
		if obj.Error == nil {
			continue
		}
		diags.AddError(
			"Failed to update data view namespaces",
			fmt.Sprintf(
				"object %s (%s): statusCode=%d error=%q message=%q",
				obj.ID,
				obj.Type,
				obj.Error.StatusCode,
				obj.Error.Error,
				obj.Error.Message,
			),
		)
	}

	return diags
}

func responseBodySnippet(body []byte) string {
	const maxLen = 1024
	if len(body) <= maxLen {
		return string(body)
	}
	return string(body[:maxLen]) + "..."
}
