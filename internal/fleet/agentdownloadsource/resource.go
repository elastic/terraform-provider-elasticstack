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

package agentdownloadsource

import (
	"context"
	"fmt"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &Resource{}
	_ resource.ResourceWithConfigure   = &Resource{}
	_ resource.ResourceWithImportState = &Resource{}
)

// Resource implements the Fleet Agent Download Source resource.
type Resource struct {
	client *clients.APIClient
}

// NewResource is a helper function to simplify the provider implementation.
func NewResource() resource.Resource {
	return &Resource{}
}

func (r *Resource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	client, diags := clients.ConvertProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	r.client = client
}

func (r *Resource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_%s", req.ProviderTypeName, "fleet_agent_download_source")
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	spaceID := "default"
	sourceID := req.ID

	if parts := strings.SplitN(req.ID, "/", 2); len(parts) == 2 {
		spaceID = parts[0]
		sourceID = parts[1]
	}

	if sourceID == "" {
		resp.Diagnostics.AddError("Invalid import identifier", "Expected <space_id>/<source_id> or <source_id>.")
		return
	}

	resp.Diagnostics.Append(setImportStateAttributes(ctx, resp, spaceID, sourceID)...)
}

func setImportStateAttributes(ctx context.Context, resp *resource.ImportStateResponse, spaceID, sourceID string) diag.Diagnostics {
	var diags diag.Diagnostics

	diags.Append(resp.State.SetAttribute(ctx, path.Root("id"), sourceID)...)
	diags.Append(resp.State.SetAttribute(ctx, path.Root("source_id"), sourceID)...)

	spaceSet, setDiags := types.SetValue(types.StringType, []attr.Value{types.StringValue(spaceID)})
	diags.Append(setDiags...)
	if diags.HasError() {
		return diags
	}
	diags.Append(resp.State.SetAttribute(ctx, path.Root("space_ids"), spaceSet)...)

	return diags
}
