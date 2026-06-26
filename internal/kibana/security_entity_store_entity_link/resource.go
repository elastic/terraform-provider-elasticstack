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

package security_entity_store_entity_link

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                   = (*EntityLinkResource)(nil)
	_ resource.ResourceWithConfigure      = (*EntityLinkResource)(nil)
	_ resource.ResourceWithImportState    = (*EntityLinkResource)(nil)
	_ resource.ResourceWithValidateConfig = (*EntityLinkResource)(nil)
)

type EntityLinkResource struct {
	*entitycore.KibanaResource[entityLinkModel]
}

// NewEntityLinkResource returns the resource.Resource for use in provider registration.
func NewEntityLinkResource() resource.Resource {
	return &EntityLinkResource{
		KibanaResource: entitycore.NewKibanaResource[entityLinkModel](
			entitycore.ComponentKibana,
			"security_entity_store_entity_link",
			entitycore.KibanaResourceOptions[entityLinkModel]{
				Schema: getResourceSchema,
				Read:   readEntityLink,
				Delete: deleteEntityLink,
				Create: createEntityLink,
				Update: updateEntityLink,
			},
		),
	}
}

// ImportState handles space-aware composite import IDs in the format
// `<space_id>/<target_id>`.
func (r *EntityLinkResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	composite, diags := clients.CompositeIDFromStr(req.ID)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(attrSpaceID), composite.ClusterID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(attrTargetID), composite.ResourceID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(attrID), types.StringValue(req.ID))...)
}
