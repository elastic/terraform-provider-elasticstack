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

package entities

import (
	"context"
	"encoding/json"
	"fmt"

	kbapi "github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	entity "github.com/elastic/terraform-provider-elasticstack/internal/kibana/security_entity_store/entity"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	jsontypes "github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	defaultSpaceID = "default"
)

func normalizeSpaceID(v types.String) string {
	if v.IsNull() || v.IsUnknown() || v.ValueString() == "" {
		return defaultSpaceID
	}
	return v.ValueString()
}

func readEntityStoreEntitiesDataSource(
	ctx context.Context,
	client *clients.KibanaScopedClient,
	model dsModel,
) (dsModel, diag.Diagnostics) {
	spaceID := normalizeSpaceID(model.SpaceID)

	params := &kbapi.GetSecurityEntityStoreEntitiesParams{}

	if !model.EntityID.IsNull() && !model.EntityID.IsUnknown() && model.EntityID.ValueString() != "" {
		filter := fmt.Sprintf(`entity.id:"%s"`, model.EntityID.ValueString())
		params.Filter = &filter
	} else if !model.Filter.IsNull() && !model.Filter.IsUnknown() {
		f := model.Filter.ValueString()
		params.Filter = &f
	}

	if !model.Size.IsNull() && !model.Size.IsUnknown() {
		s := int(model.Size.ValueInt64())
		params.Size = &s
	}
	if !model.SearchAfter.IsNull() && !model.SearchAfter.IsUnknown() {
		sa := model.SearchAfter.ValueString()
		params.SearchAfter = &sa
	}
	if !model.Source.IsNull() && !model.Source.IsUnknown() {
		src := expandStringList(model.Source)
		params.Source = &src
	}
	if !model.Fields.IsNull() && !model.Fields.IsUnknown() {
		f := expandStringList(model.Fields)
		params.Fields = &f
	}
	if !model.SortField.IsNull() && !model.SortField.IsUnknown() {
		sf := model.SortField.ValueString()
		params.SortField = &sf
	}
	if !model.SortOrder.IsNull() && !model.SortOrder.IsUnknown() {
		so := kbapi.GetSecurityEntityStoreEntitiesParamsSortOrder(model.SortOrder.ValueString())
		params.SortOrder = &so
	}
	if !model.Page.IsNull() && !model.Page.IsUnknown() {
		p := int(model.Page.ValueInt64())
		params.Page = &p
	}
	if !model.PerPage.IsNull() && !model.PerPage.IsUnknown() {
		pp := int(model.PerPage.ValueInt64())
		params.PerPage = &pp
	}
	if !model.FilterQuery.IsNull() && !model.FilterQuery.IsUnknown() {
		fq := model.FilterQuery.ValueString()
		params.FilterQuery = &fq
	}
	if !model.EntityTypes.IsNull() && !model.EntityTypes.IsUnknown() {
		types := expandEntityTypesSet(model.EntityTypes)
		params.EntityTypes = &types
	}

	resp, diags := kibanaoapi.ListSecurityEntityStoreEntities(ctx, client.GetKibanaOapiClient(), spaceID, params)
	if diags.HasError() {
		return model, diags
	}

	// Normalize JSON for results_json
	var raw any
	if err := json.Unmarshal(resp.Body, &raw); err != nil {
		return model, diag.Diagnostics{
			diag.NewErrorDiagnostic("Failed to parse response", err.Error()),
		}
	}
	normalizedBytes, err := json.Marshal(raw)
	if err != nil {
		return model, diag.Diagnostics{
			diag.NewErrorDiagnostic("Failed to normalize response", err.Error()),
		}
	}

	model.ID = types.StringValue(spaceID + "/entity_store_entities")
	model.SpaceID = types.StringValue(spaceID)
	model.ResultsJSON = jsontypes.NewNormalizedValue(string(normalizedBytes))

	// Build typed items list from the API response
	rawMap, ok := raw.(map[string]any)
	if !ok {
		return model, diag.Diagnostics{
			diag.NewErrorDiagnostic("Failed to parse response", "expected object"),
		}
	}
	var entities []any
	if rawEntities, ok := rawMap["entities"].([]any); ok {
		entities = rawEntities
	} else if rawRecords, ok := rawMap["records"].([]any); ok {
		entities = rawRecords
	}

	items := make([]attr.Value, 0, len(entities))
	for _, e := range entities {
		if doc, ok := e.(map[string]any); ok {
			item := entity.APIBodyToItem(ctx, doc, &diags)
			if diags.HasError() {
				return model, diags
			}
			items = append(items, item)
		}
	}
	itemsList, d := types.ListValue(entity.ItemObjectType(), items)
	diags.Append(d...)
	model.Items = itemsList

	return model, nil
}

func expandEntityTypesSet(s types.Set) []kbapi.GetSecurityEntityStoreEntitiesParamsEntityTypes {
	if s.IsNull() || s.IsUnknown() {
		return nil
	}
	result := make([]kbapi.GetSecurityEntityStoreEntitiesParamsEntityTypes, 0, len(s.Elements()))
	for _, v := range s.Elements() {
		if str, ok := v.(types.String); ok {
			result = append(result, kbapi.GetSecurityEntityStoreEntitiesParamsEntityTypes(str.ValueString()))
		}
	}
	return result
}
