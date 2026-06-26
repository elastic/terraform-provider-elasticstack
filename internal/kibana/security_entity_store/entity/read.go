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
	"encoding/json"
	"fmt"

	kbapi "github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	jsontypes "github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func readEntity(
	ctx context.Context,
	client *clients.KibanaScopedClient,
	_ string,
	spaceID string,
	model tfModel,
) (tfModel, bool, diag.Diagnostics) {
	spaceID = NormalizeSpaceID(types.StringValue(spaceID))
	entityID := model.EntityID.ValueString()
	entityType := model.EntityType.ValueString()

	// KQL filter strategy: entity.id:"<entityID>"
	// Kibana Query Language (KQL) uses double quotes for exact value matching.
	// The entity.id field supports free-text tokens but the quoted form treats
	// the entire string (including colons and slashes) as a single literal,
	// which is correct for all valid entity ID formats.
	// See: https://www.elastic.co/guide/en/kibana/current/kuery-query.html
	// Note: entity_types cannot be combined with KQL filter — the API treats
	// entity_types as a page-mode parameter. The filter is sufficient for
	// single-entity lookup.
	filter := fmt.Sprintf(`entity.id:%s`, QuoteKQLString(entityID))
	params := &kbapi.GetSecurityEntityStoreEntitiesParams{
		Filter: &filter,
	}

	resp, diags := kibanaoapi.ListSecurityEntityStoreEntities(ctx, client.GetKibanaOapiClient(), spaceID, params)
	if diags.HasError() {
		return model, true, diags
	}

	var result map[string]any
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return model, true, diag.Diagnostics{
			diag.NewErrorDiagnostic("Failed to parse list response", err.Error()),
		}
	}

	entities := ExtractEntitiesFromResponse(result)

	if len(entities) == 0 {
		return model, false, nil
	}

	if len(entities) > 1 {
		return model, true, diag.Diagnostics{
			diag.NewErrorDiagnostic("Ambiguous entity lookup", fmt.Sprintf("Expected exactly one entity with id %s, found %d", entityID, len(entities))),
		}
	}

	entityDoc, ok := entities[0].(map[string]any)
	if !ok {
		return model, true, diag.Diagnostics{
			diag.NewErrorDiagnostic("Invalid entity document", "Expected entity document to be an object"),
		}
	}

	model.ID = types.StringValue((&clients.CompositeID{ClusterID: spaceID, ResourceID: entityID}).String())
	model.SpaceID = types.StringValue(spaceID)
	model.EntityType = types.StringValue(entityType)
	model.EntityID = types.StringValue(entityID)
	model.ResponseJSON = jsontypes.NewNormalizedValue(string(resp.Body))

	apiBodyToModel(ctx, entityDoc, &model, &diags)
	return model, true, diags
}
