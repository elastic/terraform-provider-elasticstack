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
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/agentbuilder"
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

func (model entityLinkModel) GetVersionRequirements() ([]entitycore.VersionRequirement, diag.Diagnostics) {
	return []entitycore.VersionRequirement{
		{
			MinVersion:   *minKibanaEntityStoreResolutionVersion,
			ErrorMessage: fmt.Sprintf("Security Entity Store resolution links require Elastic Stack v%s or later.", minKibanaEntityStoreResolutionVersion),
		},
	}, nil
}

// populateFromAPI parses the raw resolution group response body and updates the
// model.  expectedEntityIDs is the set of IDs the resource manages; if any are
// absent from the API response a warning diagnostic is emitted.
func (model *entityLinkModel) populateFromAPI(ctx context.Context, spaceID string, body []byte, expectedEntityIDs []string) diag.Diagnostics {
	var diags diag.Diagnostics

	model.ID = types.StringValue((&clients.CompositeID{ClusterID: spaceID, ResourceID: model.TargetID.ValueString()}).String())
	model.SpaceID = types.StringValue(spaceID)

	// Normalize/stabilise the raw JSON before storing it.
	var rawPayload map[string]any
	if err := json.Unmarshal(body, &rawPayload); err != nil {
		diags.AddError("Failed to parse resolution group response", err.Error())
		return diags
	}
	normalised, err := json.Marshal(rawPayload)
	if err != nil {
		diags.AddError("Failed to normalise resolution group JSON", err.Error())
		return diags
	}
	model.ResolutionGroupJSON = jsontypes.NewNormalizedValue(string(normalised))

	// Extract the entity identifiers present in the resolution group.
	apiEntityIDs := extractEntityIDsFromPayload(rawPayload, model.TargetID.ValueString())

	diags.Append(agentbuilder.PopulateSet(ctx, apiEntityIDs, &model.EntityIDs)...)
	if diags.HasError() {
		return diags
	}

	// Warn when any managed entity_ids are absent from the response.
	if len(expectedEntityIDs) > 0 {
		apiSet := make(map[string]struct{}, len(apiEntityIDs))
		for _, id := range apiEntityIDs {
			apiSet[id] = struct{}{}
		}
		var missing []string
		for _, id := range expectedEntityIDs {
			if _, ok := apiSet[id]; !ok {
				missing = append(missing, id)
			}
		}
		if len(missing) > 0 {
			diags.AddWarning(
				"Missing entity IDs in resolution group",
				fmt.Sprintf("The following managed entity IDs are not present in the API response and may have been removed out-of-band: %v", missing),
			)
		}
	}

	return diags
}

// extractEntityIDsFromPayload attempts to read the list of entity identifiers
// from the raw API payload.  It looks for an "entity_ids" array and, if the
// target_id is present in that array, removes it so the resulting set contains
// only the linked alias entities.
func extractEntityIDsFromPayload(payload map[string]any, targetID string) []string {
	var result []string

	raw, ok := payload["entity_ids"]
	if !ok {
		return result
	}

	arr, ok := raw.([]any)
	if !ok {
		return result
	}

	for _, v := range arr {
		if s, ok := v.(string); ok && s != targetID {
			result = append(result, s)
		}
	}

	return result
}
