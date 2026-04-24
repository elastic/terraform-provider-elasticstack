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

package alertingrule

import (
	"context"
	_ "embed"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/resourcecore"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                   = newResource()
	_ resource.ResourceWithConfigure      = newResource()
	_ resource.ResourceWithImportState    = newResource()
	_ resource.ResourceWithValidateConfig = newResource()
	_ resource.ResourceWithUpgradeState   = newResource()
)

//go:embed resource-description.md
var resourceDescription string

type Resource struct {
	*resourcecore.Core
}

func newResource() *Resource {
	return &Resource{
		Core: resourcecore.New(resourcecore.ComponentKibana, "alerting_rule"),
	}
}

// NewResource is a helper function to simplify the provider implementation.
func NewResource() resource.Resource {
	return newResource()
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	composite, diags := clients.CompositeIDFromStrFw(req.ID)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("rule_id"), composite.ResourceID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("space_id"), composite.ClusterID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.StringValue(req.ID))...)
}
