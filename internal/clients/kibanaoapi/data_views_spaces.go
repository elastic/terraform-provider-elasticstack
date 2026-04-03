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
	"slices"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

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
			Id   string `json:"id"`
			Type string `json:"type"`
		}{
			{Id: dataViewID, Type: "index-pattern"},
		},
		SpacesToAdd:    spacesToAdd,
		SpacesToRemove: spacesToRemove,
	}

	resp, err := client.API.PostSpacesUpdateObjectsSpacesWithResponse(ctx, reqBody)
	if err != nil {
		diags.AddError("Failed to update data view namespaces", err.Error())
		return diags
	}

	if resp.StatusCode() != http.StatusOK {
		diags.AddError(
			"Unexpected status when updating data view namespaces",
			string(resp.Body),
		)
	}

	return diags
}
