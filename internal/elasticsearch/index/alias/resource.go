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

package alias

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &aliasResource{}
var _ resource.ResourceWithConfigure = &aliasResource{}
var _ resource.ResourceWithImportState = &aliasResource{}
var _ resource.ResourceWithValidateConfig = &aliasResource{}

func NewAliasResource() resource.Resource {
	return &aliasResource{}
}

type aliasResource struct {
	client *clients.APIClient
}

func (r *aliasResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_elasticsearch_index_alias"
}

func (r *aliasResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	client, diags := clients.ConvertProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	r.client = client
}

func (r *aliasResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *aliasResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config tfModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate that write_index doesn't appear in read_indices
	if config.WriteIndex.IsNull() || config.WriteIndex.IsUnknown() {
		return
	}

	if config.ReadIndices.IsNull() || config.ReadIndices.IsUnknown() {
		return
	}

	// Get the write index name
	var writeIndex indexModel
	diags := config.WriteIndex.As(ctx, &writeIndex, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if writeIndex.Name.IsUnknown() {
		return
	}
	writeIndexName := writeIndex.Name.ValueString()

	// Only validate if write index name is not empty
	if writeIndexName == "" {
		return
	}

	// Get all read indices
	var readIndices []indexModel
	if diags := config.ReadIndices.ElementsAs(ctx, &readIndices, false); !diags.HasError() {
		for _, readIndex := range readIndices {
			if readIndex.Name.IsUnknown() {
				continue
			}
			readIndexName := readIndex.Name.ValueString()
			if readIndexName != "" && readIndexName == writeIndexName {
				resp.Diagnostics.AddError(
					"Invalid Configuration",
					fmt.Sprintf("Index '%s' cannot be both a write index and a read index", writeIndexName),
				)
				return
			}
		}
	}
}
