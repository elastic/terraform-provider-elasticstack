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
	"strconv"

	jsontypes "github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ItemObjectType returns the object type for items in the list data source.
// It covers the fields that apiBodyToModel populates from an API response.
func ItemObjectType() attr.Type {
	return types.ObjectType{AttrTypes: ItemAttrTypes()}
}

// APIBodyToItem converts a raw entity document from the API list response
// into a types.Object suitable for the items list. It fills the same typed
// attributes that the resource model would have after a read.
func APIBodyToItem(ctx context.Context, body map[string]any, diags *diag.Diagnostics) types.Object {
	var item tfModel
	apiBodyToModel(ctx, body, &item, diags)
	if diags.HasError() {
		return types.ObjectNull(ItemAttrTypes())
	}
	obj, d := types.ObjectValue(ItemAttrTypes(), map[string]attr.Value{
		"@timestamp":     item.Timestamp,
		attrEntity:       item.Entity,
		attrHost:         item.Host,
		"user":           item.User,
		"service":        item.Service,
		"cloud":          item.Cloud,
		"asset":          item.Asset,
		"orchestrator":   item.Orchestrator,
		"event":          item.Event,
		attrLabels:       item.Labels,
		attrTags:         item.Tags,
		attrDocumentJSON: item.DocumentJSON,
	})
	diags.Append(d...)
	return obj
}

// ItemModel is the struct used by the data source items list.
type ItemModel = map[string]attr.Value

// ItemAttrTypes returns the attribute types for items in the list data source.
func ItemAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		attrTimestamp:    types.StringType,
		attrEntity:       types.ObjectType{AttrTypes: BlockAttrTypes()},
		attrHost:         types.ObjectType{AttrTypes: HostBlockAttrTypes()},
		attrUser:         types.ObjectType{AttrTypes: UserBlockAttrTypes()},
		attrService:      types.ObjectType{AttrTypes: ServiceBlockAttrTypes()},
		attrCloud:        types.ObjectType{AttrTypes: CloudBlockAttrTypes()},
		attrAsset:        types.ObjectType{AttrTypes: AssetBlockAttrTypes()},
		attrOrchestrator: types.ObjectType{AttrTypes: OrchestratorBlockAttrTypes()},
		attrEvent:        types.ObjectType{AttrTypes: EventBlockAttrTypes()},
		attrLabels:       types.MapType{ElemType: types.StringType},
		attrTags:         types.SetType{ElemType: types.StringType},
		attrDocumentJSON: jsontypes.NormalizedType{},
	}
}

// QuoteKQLString escapes and quotes a value for use as a KQL string literal.
// This safely handles entity IDs that may contain quotes or backslashes.
func QuoteKQLString(v string) string {
	return strconv.Quote(v)
}

// ExtractEntitiesFromResponse extracts the entity list from an API response map,
// trying "entities" first and falling back to "records" for older API versions.
func ExtractEntitiesFromResponse(result map[string]any) []any {
	if rawEntities, ok := result["entities"].([]any); ok {
		return rawEntities
	}
	if rawRecords, ok := result["records"].([]any); ok {
		return rawRecords
	}
	return nil
}
