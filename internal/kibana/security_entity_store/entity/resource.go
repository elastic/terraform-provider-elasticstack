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

package entity

import (
	"context"
	"fmt"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var (
	_ resource.ResourceWithValidateConfig = newResource()
	_ resource.ResourceWithConfigure      = newResource()
	_ resource.ResourceWithImportState    = newResource()
)

type Resource struct {
	*entitycore.KibanaResource[tfModel]
}

func newResource() *Resource {
	return &Resource{
		KibanaResource: entitycore.NewKibanaResource[tfModel](
			entitycore.ComponentKibana,
			"security_entity_store_entity",
			entitycore.KibanaResourceOptions[tfModel]{
				Schema: getSchema,
				Create: writeEntity,
				Read:   readEntity,
				Update: writeEntity,
				Delete: deleteEntity,
			},
		),
	}
}

func NewResource() resource.Resource {
	return newResource()
}

func NormalizeSpaceID(v types.String) string {
	if v.IsNull() || v.IsUnknown() || v.ValueString() == "" {
		return clients.DefaultSpaceID
	}
	return v.ValueString()
}

func buildID(spaceID, entityID string) string {
	if spaceID == "" {
		spaceID = clients.DefaultSpaceID
	}
	return fmt.Sprintf("%s/%s", spaceID, entityID)
}

// ValidateConfig implements resource.ResourceWithValidateConfig.
func (r *Resource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var model tfModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	entityID := model.EntityID.ValueString()
	if entityID == "" {
		resp.Diagnostics.AddError("Missing entity_id", "entity_id is required")
		return
	}

	if !model.Entity.IsNull() && !model.Entity.IsUnknown() {
		var entityModel entityBlockModel
		resp.Diagnostics.Append(model.Entity.As(ctx, &entityModel, basetypes.ObjectAsOptions{})...)
		if resp.Diagnostics.HasError() {
			return
		}
		if !entityModel.ID.IsNull() && !entityModel.ID.IsUnknown() && entityModel.ID.ValueString() != entityID {
			resp.Diagnostics.AddAttributeError(
				path.Root("entity_id"),
				"entity_id mismatch",
				"entity_id must match the id field in the entity block",
			)
		}
	}

	if typeutils.IsKnown(model.EntityJSON) {
		parsed := typeutils.NormalizedTypeToMap[any](model.EntityJSON, path.Root("entity_json"), &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
		if id, ok := parsed["id"].(string); ok && id != entityID {
			resp.Diagnostics.AddAttributeError(
				path.Root("entity_id"),
				"entity_id mismatch",
				"entity_id must match the id field in entity_json",
			)
		}
	}
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			"Import ID must be in the format <space_id>/<entity_id>",
		)
		return
	}
	spaceID := parts[0]
	entityID := parts[1]
	if spaceID == "" {
		spaceID = clients.DefaultSpaceID
	}
	// Derive entity_type from entity ID prefix (e.g., "host:web-01" -> "host")
	entityType := ""
	if idx := strings.Index(entityID, ":"); idx > 0 {
		entityType = entityID[:idx]
	} else {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			fmt.Sprintf("Entity ID %q must contain a type prefix (e.g., \"host:web-01\").", entityID),
		)
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("space_id"), spaceID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("entity_id"), entityID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("entity_type"), entityType)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}
