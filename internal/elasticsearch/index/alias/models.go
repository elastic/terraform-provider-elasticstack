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

package alias

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	esTypes "github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type tfModel struct {
	ID                      types.String `tfsdk:"id"`
	Name                    types.String `tfsdk:"name"`
	ElasticsearchConnection types.List   `tfsdk:"elasticsearch_connection"`
	WriteIndex              types.Object `tfsdk:"write_index"`
	ReadIndices             types.Set    `tfsdk:"read_indices"`
}

func (model tfModel) GetID() types.String                    { return model.ID }
func (model tfModel) GetResourceID() types.String            { return model.Name }
func (model tfModel) GetElasticsearchConnection() types.List { return model.ElasticsearchConnection }

func (model *tfModel) Validate(ctx context.Context) diag.Diagnostics {
	// Validate that write_index doesn't appear in read_indices.
	// This can be called during plan-time validation (unknown values possible) and during apply (typically known).

	if model.WriteIndex.IsNull() || model.WriteIndex.IsUnknown() {
		return nil
	}

	if model.ReadIndices.IsNull() || model.ReadIndices.IsUnknown() {
		return nil
	}

	// Decode write index
	var writeIndex indexModel
	diags := model.WriteIndex.As(ctx, &writeIndex, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return diags
	}

	if writeIndex.Name.IsNull() || writeIndex.Name.IsUnknown() {
		return nil
	}
	writeIndexName := writeIndex.Name.ValueString()
	if writeIndexName == "" {
		return nil
	}

	// Decode read indices and compare
	var readIndices []indexModel
	diags = model.ReadIndices.ElementsAs(ctx, &readIndices, false)
	if diags.HasError() {
		return diags
	}

	for _, readIndex := range readIndices {
		if readIndex.Name.IsNull() || readIndex.Name.IsUnknown() {
			continue
		}
		readIndexName := readIndex.Name.ValueString()
		if readIndexName != "" && readIndexName == writeIndexName {
			return diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Invalid Configuration",
					fmt.Sprintf("Index '%s' cannot be both a write index and a read index", writeIndexName),
				),
			}
		}
	}

	return nil
}

type indexModel struct {
	Name          types.String         `tfsdk:"name"`
	Filter        jsontypes.Normalized `tfsdk:"filter"`
	IndexRouting  types.String         `tfsdk:"index_routing"`
	IsHidden      types.Bool           `tfsdk:"is_hidden"`
	Routing       types.String         `tfsdk:"routing"`
	SearchRouting types.String         `tfsdk:"search_routing"`
}

// IndexConfig represents a single index configuration within an alias
type IndexConfig struct {
	Name          string
	IsWriteIndex  bool
	Filter        map[string]any
	IndexRouting  string
	IsHidden      bool
	Routing       string
	SearchRouting string
}

func (a IndexConfig) Equals(b IndexConfig) bool {
	return a.Name == b.Name &&
		a.IsWriteIndex == b.IsWriteIndex &&
		a.IndexRouting == b.IndexRouting &&
		a.IsHidden == b.IsHidden &&
		a.Routing == b.Routing &&
		a.SearchRouting == b.SearchRouting &&
		reflect.DeepEqual(a.Filter, b.Filter)
}

// aliasDefinitionToConfig converts an Elasticsearch API AliasDefinition directly to an IndexConfig.
func aliasDefinitionToConfig(indexName string, aliasData esTypes.AliasDefinition) (IndexConfig, diag.Diagnostics) {
	config := IndexConfig{
		Name:         indexName,
		IsWriteIndex: aliasData.IsWriteIndex != nil && *aliasData.IsWriteIndex,
		IsHidden:     aliasData.IsHidden != nil && *aliasData.IsHidden,
	}

	if aliasData.IndexRouting != nil {
		config.IndexRouting = *aliasData.IndexRouting
	}
	if aliasData.Routing != nil {
		config.Routing = *aliasData.Routing
	}
	if aliasData.SearchRouting != nil {
		config.SearchRouting = *aliasData.SearchRouting
	}
	if aliasData.Filter != nil {
		filterBytes, err := json.Marshal(aliasData.Filter)
		if err != nil {
			return IndexConfig{}, diag.Diagnostics{
				diag.NewErrorDiagnostic("failed to marshal alias filter", err.Error()),
			}
		}
		var filterMap map[string]any
		if err := json.Unmarshal(filterBytes, &filterMap); err != nil {
			return IndexConfig{}, diag.Diagnostics{
				diag.NewErrorDiagnostic("failed to unmarshal alias filter", err.Error()),
			}
		}
		normalized := elasticsearch.NormalizeQueryFilter(filterMap)
		if nm, ok := normalized.(map[string]any); ok {
			filterMap = nm
		}
		config.Filter = filterMap
	}

	return config, nil
}

func (model *tfModel) populateFromAPI(ctx context.Context, aliasName string, indices map[string]esTypes.AliasDefinition) diag.Diagnostics {
	model.Name = types.StringValue(aliasName)

	var writeIndex *indexModel
	var readIndices []indexModel

	for indexName, aliasData := range indices {
		// Convert AliasDefinition to indexModel
		index, err := indexFromAlias(indexName, aliasData)
		if err != nil {
			return err
		}

		if aliasData.IsWriteIndex != nil && *aliasData.IsWriteIndex {
			writeIndex = &index
		} else {
			readIndices = append(readIndices, index)
		}
	}

	// Set write index
	if writeIndex != nil {
		writeIndexObj, diags := types.ObjectValueFrom(ctx, getIndexAttrTypes(ctx), *writeIndex)
		if diags.HasError() {
			return diags
		}
		model.WriteIndex = writeIndexObj
	} else {
		model.WriteIndex = types.ObjectNull(getIndexAttrTypes(ctx))
	}

	// Set read indices
	readIndicesSet, diags := types.SetValueFrom(ctx, types.ObjectType{
		AttrTypes: getIndexAttrTypes(ctx),
	}, readIndices)
	if diags.HasError() {
		return diags
	}
	model.ReadIndices = readIndicesSet

	return nil
}

// indexFromAlias converts a esTypes.AliasDefinition to an indexModel
func indexFromAlias(indexName string, aliasData esTypes.AliasDefinition) (indexModel, diag.Diagnostics) {
	index := indexModel{
		Name:          types.StringValue(indexName),
		IsHidden:      types.BoolValue(aliasData.IsHidden != nil && *aliasData.IsHidden),
		IndexRouting:  typeutils.NonEmptyStringishPointerValue(aliasData.IndexRouting),
		Routing:       typeutils.NonEmptyStringishPointerValue(aliasData.Routing),
		SearchRouting: typeutils.NonEmptyStringishPointerValue(aliasData.SearchRouting),
	}

	if aliasData.Filter != nil {
		filterBytes, err := json.Marshal(aliasData.Filter)
		if err != nil {
			return indexModel{}, diag.Diagnostics{
				diag.NewErrorDiagnostic("failed to marshal alias filter", err.Error()),
			}
		}
		var filterMap map[string]any
		if err := json.Unmarshal(filterBytes, &filterMap); err != nil {
			return indexModel{}, diag.Diagnostics{
				diag.NewErrorDiagnostic("failed to unmarshal alias filter", err.Error()),
			}
		}
		normalized := elasticsearch.NormalizeQueryFilter(filterMap)
		if nm, ok := normalized.(map[string]any); ok {
			filterMap = nm
		}
		normalizedBytes, _ := json.Marshal(filterMap)
		index.Filter = jsontypes.NewNormalizedValue(string(normalizedBytes))
	}

	return index, nil
}

func (model *tfModel) toAliasConfigs(ctx context.Context) ([]IndexConfig, diag.Diagnostics) {
	var configs []IndexConfig

	// Handle write index
	if model.WriteIndex.IsUnknown() {
		return nil, diag.Diagnostics{
			diag.NewErrorDiagnostic(
				"Invalid Configuration",
				"Cannot build alias actions because `write_index` is unknown. Ensure `write_index` is fully known during apply.",
			),
		}
	}
	if !model.WriteIndex.IsNull() {
		var writeIndex indexModel
		diags := model.WriteIndex.As(ctx, &writeIndex, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			return nil, diags
		}
		if writeIndex.Name.IsUnknown() || writeIndex.Name.IsNull() || writeIndex.Name.ValueString() == "" {
			return nil, diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Invalid Configuration",
					"Cannot build alias actions because `write_index.name` is unknown or empty. Ensure `write_index.name` is fully known during apply.",
				),
			}
		}

		config, configDiags := indexToConfig(writeIndex, true)
		if configDiags.HasError() {
			return nil, configDiags
		}
		configs = append(configs, config)
	}

	// Handle read indices
	if model.ReadIndices.IsUnknown() {
		return nil, diag.Diagnostics{
			diag.NewErrorDiagnostic(
				"Invalid Configuration",
				"Cannot build alias actions because `read_indices` is unknown. Ensure `read_indices` is fully known during apply.",
			),
		}
	}
	if !model.ReadIndices.IsNull() {
		var readIndices []indexModel
		diags := model.ReadIndices.ElementsAs(ctx, &readIndices, false)
		if diags.HasError() {
			return nil, diags
		}

		for _, readIndex := range readIndices {
			if readIndex.Name.IsUnknown() || readIndex.Name.IsNull() || readIndex.Name.ValueString() == "" {
				return nil, diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Invalid Configuration",
						"Cannot build alias actions because one of the `read_indices` has an unknown or empty `name`. Ensure all read index names are fully known during apply.",
					),
				}
			}
			config, configDiags := indexToConfig(readIndex, false)
			if configDiags.HasError() {
				return nil, configDiags
			}
			configs = append(configs, config)
		}
	}

	return configs, nil
}

// indexToConfig converts an indexModel to IndexConfig
func indexToConfig(index indexModel, isWriteIndex bool) (IndexConfig, diag.Diagnostics) {
	config := IndexConfig{
		Name:         index.Name.ValueString(),
		IsWriteIndex: isWriteIndex,
		IsHidden:     index.IsHidden.ValueBool(),
	}

	if !index.IndexRouting.IsNull() {
		config.IndexRouting = index.IndexRouting.ValueString()
	}
	if !index.Routing.IsNull() {
		config.Routing = index.Routing.ValueString()
	}
	if !index.SearchRouting.IsNull() {
		config.SearchRouting = index.SearchRouting.ValueString()
	}
	if !index.Filter.IsNull() {
		if diags := index.Filter.Unmarshal(&config.Filter); diags.HasError() {
			return IndexConfig{}, diags
		}
	}

	return config, nil
}
