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

package datafeedstate

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func NewMLDatafeedStateResource() resource.Resource {
	return &mlDatafeedStateResource{}
}

type mlDatafeedStateResource struct {
	client *clients.APIClient
}

func (r *mlDatafeedStateResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_elasticsearch_ml_datafeed_state"
}

func (r *mlDatafeedStateResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	client, diags := clients.ConvertProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	r.client = client
}

func (r *mlDatafeedStateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to datafeed_id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("datafeed_id"), req, resp)
}
