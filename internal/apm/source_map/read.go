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

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	readPageSize = 100
)

var minSourceMapPaginationVersion = version.Must(version.NewVersion("8.7.0"))

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

	supportsPagination, vDiags := scoped.EnforceVersionCheck(ctx, func(v *version.Version) bool {
		return v.GreaterThanOrEqual(minSourceMapPaginationVersion)
	})
	diags.Append(diagutil.FrameworkDiagsFromSDK(vDiags)...)
	if diags.HasError() {
		return nil, diags
	}

	spaceID := state.SpaceID.ValueString()
	targetID := state.ID.ValueString()

	var currentPage float32 = 1
	var page, perPage *float32
	if supportsPagination {
		pp := float32(readPageSize)
		page = &currentPage
		perPage = &pp
	}

	for {
		artifacts, lDiags := kibanaoapi.ListSourceMaps(ctx, kibana, spaceID, page, perPage)
		diags.Append(lDiags...)
		if diags.HasError() {
			return nil, diags
		}

		if artifacts == nil {
			// Empty page — artifact not found.
			return nil, diags
		}

		for _, artifact := range artifacts {
			if artifact.ID != targetID {
				continue
			}

			updated := *state // copy; preserves Sourcemap, SpaceID, KibanaConnection, etc.
			updated.ID = types.StringValue(artifact.ID)
			// Only update body-derived fields when the API returned them;
			// a nil Body leaves these as empty strings in the artifact.
			if artifact.BundleFilepath != "" {
				updated.BundleFilepath = types.StringValue(artifact.BundleFilepath)
			}
			if artifact.ServiceName != "" {
				updated.ServiceName = types.StringValue(artifact.ServiceName)
			}
			if artifact.ServiceVersion != "" {
				updated.ServiceVersion = types.StringValue(artifact.ServiceVersion)
			}
			return &updated, diags
		}

		if !supportsPagination || len(artifacts) < readPageSize {
			break
		}
		currentPage++
	}

	return nil, diags
}
