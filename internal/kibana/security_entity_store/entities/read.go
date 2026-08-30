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
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	jsontypes "github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func readEntityStoreEntitiesDataSource(
	ctx context.Context,
	client *clients.KibanaScopedClient,
	model dsModel,
) (dsModel, diag.Diagnostics) {
	spaceID := entity.NormalizeSpaceID(model.SpaceID)

	params := &kbapi.GetSecurityEntityStoreEntitiesParams{}
	var diags diag.Diagnostics

	if typeutils.IsKnown(model.EntityID) && model.EntityID.ValueString() != "" {
		filter := fmt.Sprintf(`entity.id:%s`, entity.QuoteKQLString(model.EntityID.ValueString()))
		params.Filter = &filter
	} else if typeutils.IsKnown(model.Filter) {
		f := model.Filter.ValueString()
		params.Filter = &f
	}

	if typeutils.IsKnown(model.Size) {
		params.Size = typeutils.OptionalInt(model.Size)
	}
	if typeutils.IsKnown(model.SearchAfter) {
		sa := model.SearchAfter.ValueString()
		params.SearchAfter = &sa
	}
	if typeutils.IsKnown(model.Source) {
		src := typeutils.ListTypeToSliceString(ctx, model.Source, path.Root("source"), &diags)
		if diags.HasError() {
			return model, diags
		}
		params.Source = &src
	}
	if typeutils.IsKnown(model.Fields) {
		f := typeutils.ListTypeToSliceString(ctx, model.Fields, path.Root("fields"), &diags)
		if diags.HasError() {
			return model, diags
		}
		params.Fields = &f
	}
	if typeutils.IsKnown(model.SortField) {
		sf := model.SortField.ValueString()
		params.SortField = &sf
	}
	if typeutils.IsKnown(model.SortOrder) {
		so := kbapi.GetSecurityEntityStoreEntitiesParamsSortOrder(model.SortOrder.ValueString())
		params.SortOrder = &so
	}
	if typeutils.IsKnown(model.Page) {
		params.Page = typeutils.OptionalInt(model.Page)
	}
	if typeutils.IsKnown(model.PerPage) {
		params.PerPage = typeutils.OptionalInt(model.PerPage)
	}
	if typeutils.IsKnown(model.FilterQuery) {
		fq := model.FilterQuery.ValueString()
		params.FilterQuery = &fq
	}
	if typeutils.IsKnown(model.EntityTypes) {
		entityTypes := expandEntityTypesSet(model.EntityTypes)
		params.EntityTypes = &entityTypes
	}

	var resp *kbapi.GetSecurityEntityStoreEntitiesResponse
	resp, diags = kibanaoapi.ListSecurityEntityStoreEntities(ctx, client.GetKibanaOapiClient(), spaceID, params)
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
	entities := entity.ExtractEntitiesFromResponse(rawMap)

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
	var diags diag.Diagnostics
	strs := typeutils.StringElements(s, &diags)
	result := make([]kbapi.GetSecurityEntityStoreEntitiesParamsEntityTypes, 0, len(strs))
	for _, str := range strs {
		result = append(result, kbapi.GetSecurityEntityStoreEntitiesParamsEntityTypes(str))
	}
	return result
}
