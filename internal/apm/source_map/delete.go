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

package sourcemap

import (
	"context"
	"fmt"
	"net/http"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanautil"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func (r *resourceSourceMap) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state SourceMap
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	scoped, fwDiags := r.Client().GetKibanaClient(ctx, state.KibanaConnection)
	resp.Diagnostics.Append(fwDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	kibana, err := scoped.GetKibanaOapiClient()
	if err != nil {
		resp.Diagnostics.AddError("Unable to get Kibana client", err.Error())
		return
	}

	artifactID := state.ID.ValueString()
	spaceID := state.SpaceID.ValueString()

	apiResp, err := kibana.API.DeleteSourceMapWithResponse(
		ctx,
		artifactID,
		&kbapi.DeleteSourceMapParams{
			ElasticApiVersion: kbapi.N20231031,
		},
		kibanautil.SpaceAwarePathRequestEditor(spaceID),
	)
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete APM source map", err.Error())
		return
	}

	if apiResp.HTTPResponse.StatusCode == http.StatusNotFound {
		return
	}

	if apiResp.HTTPResponse.StatusCode >= 400 {
		resp.Diagnostics.Append(diagutil.ReportUnknownHTTPError(apiResp.HTTPResponse.StatusCode, apiResp.Body)...)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("Deleted APM source map with ID: %s", artifactID))
}
