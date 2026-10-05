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
	"fmt"
	"reflect"

	esTypes "github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index/aliasutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type tfModel struct {
	entitycore.ResourceTimeoutsField
	ID                      types.String `tfsdk:"id"`
	Name                    types.String `tfsdk:"name"`
	ElasticsearchConnection types.List   `tfsdk:"elasticsearch_connection"`
	WriteIndex              types.Object `tfsdk:"write_index"`
	ReadIndices             types.Set    `tfsdk:"read_indices"`
	desiredEmptyAfterWrite  bool
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
	var readIndices []readIndexModel
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

type readIndexModel struct {
	Name            types.String         `tfsdk:"name"`
	ConcreteIndices types.Set            `tfsdk:"concrete_indices"`
	Filter          jsontypes.Normalized `tfsdk:"filter"`
	IndexRouting    types.String         `tfsdk:"index_routing"`
	IsHidden        types.Bool           `tfsdk:"is_hidden"`
	Routing         types.String         `tfsdk:"routing"`
	SearchRouting   types.String         `tfsdk:"search_routing"`
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
		filterMap, diags := aliasutil.NormalizeAliasFilterAnyToMap(aliasData.Filter)
		if diags.HasError() {
			return IndexConfig{}, diags
		}
		config.Filter = filterMap
	}

	return config, nil
}

func (model *tfModel) populateFromAPI(ctx context.Context, aliasName string, indices map[string]esTypes.AliasDefinition) diag.Diagnostics {
	model.Name = types.StringValue(aliasName)

	var writeIndex *indexModel
	var readIndices []readIndexModel

	for indexName, aliasData := range indices {
		// Convert AliasDefinition to indexModel
		index, err := indexFromAlias(indexName, aliasData)
		if err != nil {
			return err
		}

		if aliasData.IsWriteIndex != nil && *aliasData.IsWriteIndex {
			writeIndex = &indexModel{
				Name:          index.Name,
				Filter:        index.Filter,
				IndexRouting:  index.IndexRouting,
				IsHidden:      index.IsHidden,
				Routing:       index.Routing,
				SearchRouting: index.SearchRouting,
			}
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
		AttrTypes: getReadIndexAttrTypes(ctx),
	}, readIndices)
	if diags.HasError() {
		return diags
	}
	model.ReadIndices = readIndicesSet

	return nil
}

func (model *tfModel) populateReadState(
	ctx context.Context,
	aliasName string,
	indices map[string]esTypes.AliasDefinition,
	resolveIndexExpression resolveIndexExpressionFunc,
) diag.Diagnostics {
	model.Name = types.StringValue(aliasName)

	var writeIndex *indexModel
	readAliasData := make(map[string]esTypes.AliasDefinition)
	for indexName, aliasData := range indices {
		index, diags := indexFromAlias(indexName, aliasData)
		if diags.HasError() {
			return diags
		}

		if aliasData.IsWriteIndex != nil && *aliasData.IsWriteIndex {
			writeIndex = &indexModel{
				Name:          index.Name,
				Filter:        index.Filter,
				IndexRouting:  index.IndexRouting,
				IsHidden:      index.IsHidden,
				Routing:       index.Routing,
				SearchRouting: index.SearchRouting,
			}
			continue
		}

		readAliasData[indexName] = aliasData
	}

	if writeIndex != nil {
		writeIndexObj, diags := types.ObjectValueFrom(ctx, getIndexAttrTypes(ctx), *writeIndex)
		if diags.HasError() {
			return diags
		}
		model.WriteIndex = writeIndexObj
	} else {
		model.WriteIndex = types.ObjectNull(getIndexAttrTypes(ctx))
	}

	var priorReadIndices []readIndexModel
	if !model.ReadIndices.IsNull() && !model.ReadIndices.IsUnknown() {
		diags := model.ReadIndices.ElementsAs(ctx, &priorReadIndices, false)
		if diags.HasError() {
			return diags
		}
	}

	if len(priorReadIndices) == 0 {
		return model.populateFromAPI(ctx, aliasName, indices)
	}

	covered := make(map[string]struct{})
	readIndices := make([]readIndexModel, 0, len(priorReadIndices)+len(readAliasData))
	for _, readIndex := range priorReadIndices {
		targets, diags := resolveIndexExpression(ctx, readIndex.Name.ValueString())
		if diags.HasError() {
			return diags
		}

		concreteIndices := make([]attr.Value, 0, len(targets.Names))
		for _, target := range targets.Names {
			if _, exists := readAliasData[target]; !exists {
				continue
			}
			covered[target] = struct{}{}
			concreteIndices = append(concreteIndices, types.StringValue(target))
		}
		readIndex.ConcreteIndices = types.SetValueMust(types.StringType, concreteIndices)
		readIndices = append(readIndices, readIndex)
	}

	for indexName, aliasData := range readAliasData {
		if _, exists := covered[indexName]; exists {
			continue
		}
		readIndex, diags := indexFromAlias(indexName, aliasData)
		if diags.HasError() {
			return diags
		}
		readIndices = append(readIndices, readIndex)
	}

	readIndicesSet, diags := types.SetValueFrom(ctx, types.ObjectType{
		AttrTypes: getReadIndexAttrTypes(ctx),
	}, readIndices)
	if diags.HasError() {
		return diags
	}
	model.ReadIndices = readIndicesSet

	return nil
}

func (model tfModel) isVirtualState(ctx context.Context) bool {
	if !model.WriteIndex.IsNull() {
		return false
	}
	if model.ReadIndices.IsNull() {
		return true
	}
	if model.ReadIndices.IsUnknown() {
		return false
	}

	var readIndices []readIndexModel
	diags := model.ReadIndices.ElementsAs(ctx, &readIndices, false)
	if diags.HasError() {
		return false
	}
	for _, readIndex := range readIndices {
		if readIndex.ConcreteIndices.IsNull() || readIndex.ConcreteIndices.IsUnknown() {
			return false
		}
		var concreteIndices []string
		diags := readIndex.ConcreteIndices.ElementsAs(ctx, &concreteIndices, false)
		if diags.HasError() || len(concreteIndices) > 0 {
			return false
		}
	}

	return true
}

func (model *tfModel) markDesiredEmptyAfterWrite(ctx context.Context) diag.Diagnostics {
	if model.ReadIndices.IsNull() {
		model.desiredEmptyAfterWrite = true
		return nil
	}

	var readIndices []readIndexModel
	diags := model.ReadIndices.ElementsAs(ctx, &readIndices, false)
	if diags.HasError() {
		return diags
	}
	for index := range readIndices {
		readIndices[index].ConcreteIndices = types.SetValueMust(types.StringType, nil)
	}

	readIndicesSet, diags := types.SetValueFrom(ctx, types.ObjectType{
		AttrTypes: getReadIndexAttrTypes(ctx),
	}, readIndices)
	if diags.HasError() {
		return diags
	}
	model.ReadIndices = readIndicesSet
	model.desiredEmptyAfterWrite = true

	return nil
}

// indexFromAlias converts an esTypes.AliasDefinition to a readIndexModel
func indexFromAlias(indexName string, aliasData esTypes.AliasDefinition) (readIndexModel, diag.Diagnostics) {
	index := readIndexModel{
		Name:            types.StringValue(indexName),
		ConcreteIndices: types.SetValueMust(types.StringType, []attr.Value{types.StringValue(indexName)}),
		IsHidden:        types.BoolValue(aliasData.IsHidden != nil && *aliasData.IsHidden),
		IndexRouting:    typeutils.NonEmptyStringishPointerValue(aliasData.IndexRouting),
		Routing:         typeutils.NonEmptyStringishPointerValue(aliasData.Routing),
		SearchRouting:   typeutils.NonEmptyStringishPointerValue(aliasData.SearchRouting),
	}

	filter, diags := aliasutil.NormalizeAliasFilterFromAny(aliasData.Filter)
	if diags.HasError() {
		return readIndexModel{}, diags
	}
	index.Filter = filter

	return index, nil
}

type resolveIndexExpressionFunc func(context.Context, string) (elasticsearch.ResolvedIndexTargets, diag.Diagnostics)

func (model *tfModel) resolveAliasConfigs(ctx context.Context, resolveIndexExpression resolveIndexExpressionFunc) ([]IndexConfig, diag.Diagnostics) {
	var configs []IndexConfig
	var writeIndex *IndexConfig

	if model.WriteIndex.IsUnknown() {
		return nil, diag.Diagnostics{
			diag.NewErrorDiagnostic(
				"Invalid Configuration",
				"Cannot build alias actions because `write_index` is unknown. Ensure `write_index` is fully known during apply.",
			),
		}
	}
	if !model.WriteIndex.IsNull() {
		var writeIndexModel indexModel
		diags := model.WriteIndex.As(ctx, &writeIndexModel, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			return nil, diags
		}
		if writeIndexModel.Name.IsUnknown() || writeIndexModel.Name.IsNull() || writeIndexModel.Name.ValueString() == "" {
			return nil, diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Invalid Configuration",
					"Cannot build alias actions because `write_index.name` is unknown or empty. Ensure `write_index.name` is fully known during apply.",
				),
			}
		}
		if isWriteIndexSelector(writeIndexModel.Name.ValueString()) {
			return nil, diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Invalid Configuration",
					fmt.Sprintf("Write index name %q must name a single index", writeIndexModel.Name.ValueString()),
				),
			}
		}

		config, configDiags := indexToConfig(writeIndexModel, true)
		if configDiags.HasError() {
			return nil, configDiags
		}
		writeIndex = &config
		configs = append(configs, config)
	}

	if model.ReadIndices.IsNull() {
		return configs, nil
	}
	if model.ReadIndices.IsUnknown() {
		return nil, diag.Diagnostics{
			diag.NewErrorDiagnostic(
				"Invalid Configuration",
				"Cannot build alias actions because `read_indices` is unknown. Ensure `read_indices` is fully known during apply.",
			),
		}
	}

	var readIndices []readIndexModel
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
	}

	configByTarget := make(map[string]IndexConfig)
	expressionByTarget := make(map[string]string)
	var targetKind elasticsearch.IndexTargetKind
	hasTargetKind := false
	for _, readIndex := range readIndices {
		expression := readIndex.Name.ValueString()
		targets, resolveDiags := resolveIndexExpression(ctx, expression)
		if resolveDiags.HasError() {
			return nil, resolveDiags
		}
		if len(targets.Names) > 0 {
			if hasTargetKind && targetKind != targets.Kind {
				return nil, diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Invalid Configuration",
						"Read index expressions resolve to both regular indices and data streams",
					),
				}
			}
			targetKind = targets.Kind
			hasTargetKind = true
		}

		config, configDiags := readIndexToConfig(readIndex)
		if configDiags.HasError() {
			return nil, configDiags
		}

		for _, target := range targets.Names {
			if writeIndex != nil && target == writeIndex.Name {
				return nil, diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Invalid Configuration",
						fmt.Sprintf("Read index expression %q resolves to write index %q", expression, target),
					),
				}
			}

			config.Name = target
			if existing, exists := configByTarget[target]; exists {
				if !existing.Equals(config) {
					return nil, diag.Diagnostics{
						diag.NewErrorDiagnostic(
							"Invalid Configuration",
							fmt.Sprintf("Read index expressions %q and %q resolve to target %q with conflicting settings", expressionByTarget[target], expression, target),
						),
					}
				}
				continue
			}

			configByTarget[target] = config
			expressionByTarget[target] = expression
			configs = append(configs, config)
		}
	}

	return configs, nil
}

func readIndexToConfig(index readIndexModel) (IndexConfig, diag.Diagnostics) {
	return indexToConfig(indexModel{
		Name:          index.Name,
		Filter:        index.Filter,
		IndexRouting:  index.IndexRouting,
		IsHidden:      index.IsHidden,
		Routing:       index.Routing,
		SearchRouting: index.SearchRouting,
	}, false)
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

func buildAliasActions(aliasName string, current map[string]IndexConfig, desired []IndexConfig) []elasticsearch.AliasAction {
	desiredByName := make(map[string]IndexConfig, len(desired))
	for _, config := range desired {
		desiredByName[config.Name] = config
	}

	actions := make([]elasticsearch.AliasAction, 0, len(current)+len(desired))
	for name := range current {
		if _, exists := desiredByName[name]; !exists {
			actions = append(actions, elasticsearch.AliasAction{
				Type:  "remove",
				Index: name,
				Alias: aliasName,
			})
		}
	}

	for _, config := range desired {
		if currentConfig, exists := current[config.Name]; exists && currentConfig.Equals(config) {
			continue
		}
		actions = append(actions, elasticsearch.AliasAction{
			Type:          "add",
			Index:         config.Name,
			Alias:         aliasName,
			IsWriteIndex:  config.IsWriteIndex,
			Filter:        config.Filter,
			IndexRouting:  config.IndexRouting,
			IsHidden:      config.IsHidden,
			Routing:       config.Routing,
			SearchRouting: config.SearchRouting,
		})
	}

	return actions
}

func (model *tfModel) buildResolvedAliasActions(
	ctx context.Context,
	aliasName string,
	current map[string]IndexConfig,
	resolveIndexExpression resolveIndexExpressionFunc,
) ([]elasticsearch.AliasAction, diag.Diagnostics) {
	actions, _, diags := model.buildResolvedAliasActionsWithOutcome(ctx, aliasName, current, resolveIndexExpression)
	return actions, diags
}

func (model *tfModel) buildResolvedAliasActionsWithOutcome(
	ctx context.Context,
	aliasName string,
	current map[string]IndexConfig,
	resolveIndexExpression resolveIndexExpressionFunc,
) ([]elasticsearch.AliasAction, bool, diag.Diagnostics) {
	desired, diags := model.resolveAliasConfigs(ctx, resolveIndexExpression)
	if diags.HasError() {
		return nil, false, diags
	}

	return buildAliasActions(aliasName, current, desired), len(desired) == 0, nil
}

func currentAliasConfigs(aliasName string, indices map[string]esTypes.IndexAliases) (map[string]IndexConfig, diag.Diagnostics) {
	configs := make(map[string]IndexConfig)
	for indexName, indexAliases := range indices {
		aliasDefinition, exists := indexAliases.Aliases[aliasName]
		if !exists {
			continue
		}

		config, diags := aliasDefinitionToConfig(indexName, aliasDefinition)
		if diags.HasError() {
			return nil, diags
		}
		configs[indexName] = config
	}

	return configs, nil
}
