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
	"encoding/json"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ entitycore.KibanaResourceModel     = entityLinkModel{}
	_ entitycore.WithVersionRequirements = entityLinkModel{}
)

var minKibanaEntityStoreResolutionVersion = version.Must(version.NewVersion("9.1.0"))

type entityLinkModel struct {
	entitycore.ResourceTimeoutsField
	entitycore.KibanaConnectionField
	ID                  types.String         `tfsdk:"id"`
	SpaceID             types.String         `tfsdk:"space_id"`
	TargetID            types.String         `tfsdk:"target_id"`
	EntityIDs           types.Set            `tfsdk:"entity_ids"`
	ResolutionGroupJSON jsontypes.Normalized `tfsdk:"resolution_group_json"`
}

func (model entityLinkModel) GetID() types.String         { return model.ID }
func (model entityLinkModel) GetResourceID() types.String { return model.TargetID }
func (model entityLinkModel) GetSpaceID() types.String    { return model.SpaceID }

func (model entityLinkModel) GetVersionRequirements(_ context.Context) ([]entitycore.VersionRequirement, diag.Diagnostics) {
	return []entitycore.VersionRequirement{
		{
			MinVersion:   *minKibanaEntityStoreResolutionVersion,
			ErrorMessage: fmt.Sprintf("Security Entity Store resolution links require Elastic Stack v%s or later.", minKibanaEntityStoreResolutionVersion),
		},
	}, nil
}

// populateFromAPI updates the model from a parsed resolution group response.
func (model *entityLinkModel) populateFromAPI(ctx context.Context, spaceID string, payload map[string]any) diag.Diagnostics {
	var diags diag.Diagnostics

	model.ID = types.StringValue((&clients.CompositeID{ClusterID: spaceID, ResourceID: model.TargetID.ValueString()}).String())
	model.SpaceID = types.StringValue(spaceID)

	normalised, err := json.Marshal(payload)
	if err != nil {
		diags.AddError("Failed to normalise resolution group JSON", err.Error())
		return diags
	}
	model.ResolutionGroupJSON = jsontypes.NewNormalizedValue(string(normalised))

	// Extract the entity identifiers present in the resolution group.
	apiEntityIDs := extractEntityIDsFromPayload(payload, model.TargetID.ValueString())

	entityIDsSet, setDiag := typeutils.StringSetOrNull(ctx, apiEntityIDs)
	diags.Append(setDiag...)
	if setDiag.HasError() {
		return diags
	}
	model.EntityIDs = entityIDsSet

	return diags
}

// extractEntityIDsFromPayload attempts to read the list of alias entity
// identifiers from the raw API payload.  It walks the "aliases" array and
// extracts each alias's "entity.id" value, filtering out the target itself.
func extractEntityIDsFromPayload(payload map[string]any, targetID string) []string {
	var result []string

	rawAliases, ok := payload["aliases"]
	if !ok {
		return result
	}

	aliases, ok := rawAliases.([]any)
	if !ok {
		return result
	}

	for _, v := range aliases {
		aliasMap, ok := v.(map[string]any)
		if !ok {
			continue
		}
		entityRaw, ok := aliasMap["entity"]
		if !ok {
			continue
		}
		entityMap, ok := entityRaw.(map[string]any)
		if !ok {
			continue
		}
		idRaw, ok := entityMap["id"]
		if !ok {
			continue
		}
		if idStr, ok := idRaw.(string); ok && idStr != targetID {
			result = append(result, idStr)
		}
	}

	return result
}
