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

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanautil"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

const (
	readPageSize = 100
)

func (r *resourceSourceMap) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state SourceMap
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updatedState, diags := r.read(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if updatedState == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, updatedState)...)
}

// read is the internal read function shared by Read and Create.
// It reads space_id from state and paginates through the source maps list to
// find the artifact matching state.ID. Returns nil if not found (resource removed
// from state), or the updated state if found.
func (r *resourceSourceMap) read(ctx context.Context, state *SourceMap) (*SourceMap, diag.Diagnostics) {
	var diags diag.Diagnostics

	scoped, kDiags := r.Client().GetKibanaClient(ctx, state.KibanaConnection)
	diags.Append(kDiags...)
	if diags.HasError() {
		return nil, diags
	}

	kibana, err := scoped.GetKibanaOapiClient()
	if err != nil {
		diags.AddError("Unable to get Kibana client", err.Error())
		return nil, diags
	}

	spaceID := state.SpaceID.ValueString()
	targetID := state.ID.ValueString()

	page := float32(1)
	perPage := float32(readPageSize)

	for {
		apiResp, apiErr := kibana.API.GetSourceMapsWithResponse(
			ctx,
			&kbapi.GetSourceMapsParams{
				Page:              &page,
				PerPage:           &perPage,
				ElasticApiVersion: kbapi.GetSourceMapsParamsElasticApiVersionN20231031,
			},
			kibanautil.SpaceAwarePathRequestEditor(spaceID),
		)
		if apiErr != nil {
			diags.AddError("Failed to list APM source maps", apiErr.Error())
			return nil, diags
		}

		if httpDiags := diagutil.CheckHTTPErrorFromFW(apiResp.HTTPResponse, "Failed to list APM source maps"); httpDiags.HasError() {
			diags.Append(httpDiags...)
			return nil, diags
		}

		if apiResp.JSON200 == nil || apiResp.JSON200.Artifacts == nil {
			// Empty or nil body — artifact not found.
			return nil, diags
		}

		artifacts := *apiResp.JSON200.Artifacts
		for i := range artifacts {
			artifact := &artifacts[i]
			if artifact.Id == nil || *artifact.Id != targetID {
				continue
			}

			// Found the matching artifact — populate state from the response.
			updated := *state // copy
			updated.ID = typeutils.StringishPointerValue(artifact.Id)
			if artifact.Body != nil {
				updated.BundleFilepath = typeutils.StringishPointerValue(artifact.Body.BundleFilepath)
				updated.ServiceName = typeutils.StringishPointerValue(artifact.Body.ServiceName)
				updated.ServiceVersion = typeutils.StringishPointerValue(artifact.Body.ServiceVersion)
			}
			// space_id is preserved from state; the API does not return space metadata.
			// sourcemap_json and sourcemap_binary are not repopulated; the API does not
			// return the original uploaded content.
			return &updated, diags
		}

		// If the page returned fewer items than perPage, we've read the last page.
		if len(artifacts) < readPageSize {
			break
		}
		page++
	}

	// Artifact not found — signal removal from state.
	return nil, diags
}
