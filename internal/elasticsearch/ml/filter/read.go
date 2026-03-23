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

package filter

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	fwdiags "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func (r *filterResource) read(ctx context.Context, model *FilterTFModel) (bool, fwdiags.Diagnostics) {
	var diags fwdiags.Diagnostics

	if !r.resourceReady(&diags) {
		return false, diags
	}

	compID, sdkDiags := clients.CompositeIdFromStr(model.ID.ValueString())
	diags.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if diags.HasError() {
		return false, diags
	}
	filterID := compID.ResourceId

	tflog.Debug(ctx, fmt.Sprintf("Reading ML filter: %s", filterID))

	esClient, err := r.client.GetESClient()
	if err != nil {
		diags.AddError("Failed to get Elasticsearch client", err.Error())
		return false, diags
	}

	res, err := esClient.ML.GetFilters(esClient.ML.GetFilters.WithFilterID(filterID), esClient.ML.GetFilters.WithContext(ctx))
	if err != nil {
		diags.AddError("Failed to get ML filter", err.Error())
		return false, diags
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		return false, nil
	}

	getDiags := diagutil.CheckErrorFromFW(res, fmt.Sprintf("Unable to get ML filter: %s", filterID))
	diags.Append(getDiags...)
	if diags.HasError() {
		return false, diags
	}

	var response struct {
		Filters []FilterAPIModel `json:"filters"`
		Count   int              `json:"count"`
	}
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		diags.AddError("Failed to decode filter response", err.Error())
		return false, diags
	}

	if len(response.Filters) == 0 {
		return false, nil
	}

	diags.Append(model.fromAPIModel(ctx, &response.Filters[0])...)
	if diags.HasError() {
		return false, diags
	}

	tflog.Debug(ctx, fmt.Sprintf("Successfully read ML filter: %s", filterID))
	return true, diags
}
