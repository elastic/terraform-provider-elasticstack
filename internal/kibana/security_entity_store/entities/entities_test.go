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
	"reflect"
	"testing"

	kbapi "github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// hasMixedPaginationModes checks whether page-mode and cursor-mode parameters are both set.
func hasMixedPaginationModes(model dsModel) bool {
	hasPageMode := false
	if !model.SortField.IsNull() || !model.SortOrder.IsNull() || !model.Page.IsNull() || !model.PerPage.IsNull() || !model.FilterQuery.IsNull() {
		hasPageMode = true
	}
	hasCursorMode := false
	if !model.Filter.IsNull() || !model.Size.IsNull() || !model.SearchAfter.IsNull() || !model.Source.IsNull() || !model.Fields.IsNull() {
		hasCursorMode = true
	}
	return hasPageMode && hasCursorMode
}

func TestHasMixedPaginationModes(t *testing.T) {
	tests := []struct {
		name  string
		model dsModel
		want  bool
	}{
		{
			name:  "no params",
			model: dsModel{SpaceID: types.StringValue("default")},
			want:  false,
		},
		{
			name:  "page mode only",
			model: dsModel{SpaceID: types.StringValue("default"), Page: types.Int64Value(1)},
			want:  false,
		},
		{
			name:  "cursor mode only",
			model: dsModel{SpaceID: types.StringValue("default"), Filter: types.StringValue("entity.type:host")},
			want:  false,
		},
		{
			name:  "mixed modes",
			model: dsModel{SpaceID: types.StringValue("default"), Page: types.Int64Value(1), Filter: types.StringValue("entity.type:host")},
			want:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasMixedPaginationModes(tt.model)
			if got != tt.want {
				t.Errorf("hasMixedPaginationModes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExpandStringList(t *testing.T) {
	tests := []struct {
		name  string
		input types.List
		want  []string
	}{
		{
			name:  "null list",
			input: types.ListNull(types.StringType),
			want:  nil,
		},
		{
			name:  "empty list",
			input: types.ListValueMust(types.StringType, []attr.Value{}),
			want:  []string{},
		},
		{
			name:  "non-empty list",
			input: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("a"), types.StringValue("b")}),
			want:  []string{"a", "b"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := expandStringList(tt.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("expandStringList() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExpandEntityTypesSet(t *testing.T) {
	tests := []struct {
		name  string
		input types.Set
		want  []kbapi.GetSecurityEntityStoreEntitiesParamsEntityTypes
	}{
		{
			name:  "null set",
			input: types.SetNull(types.StringType),
			want:  nil,
		},
		{
			name:  "empty set",
			input: types.SetValueMust(types.StringType, []attr.Value{}),
			want:  []kbapi.GetSecurityEntityStoreEntitiesParamsEntityTypes{},
		},
		{
			name:  "host and user",
			input: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("host"), types.StringValue("user")}),
			want:  []kbapi.GetSecurityEntityStoreEntitiesParamsEntityTypes{kbapi.GetSecurityEntityStoreEntitiesParamsEntityTypes("host"), kbapi.GetSecurityEntityStoreEntitiesParamsEntityTypes("user")},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := expandEntityTypesSet(tt.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("expandEntityTypesSet() = %v, want %v", got, tt.want)
			}
		})
	}
}