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

package integration

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var (
	_ resource.Resource                 = &integrationResource{}
	_ resource.ResourceWithConfigure    = &integrationResource{}
	_ resource.ResourceWithUpgradeState = &integrationResource{}

	// MinVersionIgnoreMappingUpdateErrors is the minimum version that supports the ignore_mapping_update_errors parameter
	MinVersionIgnoreMappingUpdateErrors = version.Must(version.NewVersion("8.11.0"))
	// MinVersionSkipDataStreamRollover is the minimum version that supports the skip_data_stream_rollover parameter
	MinVersionSkipDataStreamRollover = MinVersionIgnoreMappingUpdateErrors
)

// NewResource is a helper function to simplify the provider implementation.
func NewResource() resource.Resource {
	return &integrationResource{}
}

type integrationResource struct {
	client *clients.APIClient
}

func (r *integrationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	client, diags := clients.ConvertProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	r.client = client
}

func (r *integrationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_%s", req.ProviderTypeName, "fleet_integration")
}
