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

package parameter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/synthetics"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *Resource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var plan tfModelV0
	diags := request.State.Get(ctx, &plan)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	apiClient, diags := r.client.GetKibanaClient(ctx, plan.KibanaConnection)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	kibanaClient := synthetics.GetKibanaOAPIClientFromScopedClient(apiClient, response.Diagnostics)
	if kibanaClient == nil {
		return
	}

	resourceID := plan.ID.ValueString()

	compositeID, dg := tryReadCompositeID(resourceID)
	response.Diagnostics.Append(dg...)
	if response.Diagnostics.HasError() {
		return
	}

	if compositeID != nil {
		resourceID = compositeID.ResourceID
	}

	// DELETE /api/synthetics/params/{id} only works on Kibana >= 8.17.0.
	// DELETE /api/synthetics/params with {"ids": [...]} works on >= 8.12.0.
	// We use the latter so all supported Kibana versions are covered, sending
	// the request through the kbapi transport (which handles auth and headers).
	body, err := json.Marshal(map[string][]string{"ids": {resourceID}})
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("Failed to marshal delete request for parameter `%s`", resourceID), err.Error())
		return
	}

	endpoint := strings.TrimRight(kibanaClient.URL, "/") + "/api/synthetics/params"
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, endpoint, bytes.NewReader(body))
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("Failed to build delete request for parameter `%s`", resourceID), err.Error())
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := kibanaClient.HTTP.Do(req)
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("Failed to delete parameter `%s`", resourceID), err.Error())
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		response.Diagnostics.AddError(
			fmt.Sprintf("Unexpected status deleting parameter `%s`", resourceID),
			fmt.Sprintf("API returned status %s", resp.Status),
		)
		return
	}
}
