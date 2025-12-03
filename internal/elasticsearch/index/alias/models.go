package alias

import (
	"context"
	"encoding/json"
	"reflect"

	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type tfModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	WriteIndex  types.Object `tfsdk:"write_index"`
	ReadIndices types.Set    `tfsdk:"read_indices"`
}

type indexModel struct {
	Name          types.String         `tfsdk:"name"`
	Filter        jsontypes.Normalized `tfsdk:"filter"`
	IndexRouting  types.String         `tfsdk:"index_routing"`
	IsHidden      types.Bool           `tfsdk:"is_hidden"`
	Routing       types.String         `tfsdk:"routing"`
	SearchRouting types.String         `tfsdk:"search_routing"`
}

// AliasIndexConfig represents a single index configuration within an alias
type AliasIndexConfig struct {
	Name          string
	IsWriteIndex  bool
	Filter        map[string]interface{}
	IndexRouting  string
	IsHidden      bool
	Routing       string
	SearchRouting string
}

func (a AliasIndexConfig) Equals(b AliasIndexConfig) bool {
	return a.Name == b.Name &&
		a.IsWriteIndex == b.IsWriteIndex &&
		a.IndexRouting == b.IndexRouting &&
		a.IsHidden == b.IsHidden &&
		a.Routing == b.Routing &&
		a.SearchRouting == b.SearchRouting &&
		reflect.DeepEqual(a.Filter, b.Filter)
}

func (model *tfModel) populateFromAPI(ctx context.Context, aliasName string, indices map[string]models.IndexAlias) diag.Diagnostics {
	model.Name = types.StringValue(aliasName)

	var writeIndex *indexModel
	var readIndices []indexModel

	for indexName, aliasData := range indices {
		// Convert IndexAlias to indexModel
		index, err := indexFromAlias(indexName, aliasData)
		if err != nil {
			return err
		}

		if aliasData.IsWriteIndex {
			writeIndex = &index
		} else {
			readIndices = append(readIndices, index)
		}
	}

	// Set write index
	if writeIndex != nil {
		writeIndexObj, diags := types.ObjectValueFrom(ctx, getIndexAttrTypes(), *writeIndex)
		if diags.HasError() {
			return diags
		}
		model.WriteIndex = writeIndexObj
	} else {
		model.WriteIndex = types.ObjectNull(getIndexAttrTypes())
	}

	// Set read indices
	readIndicesSet, diags := types.SetValueFrom(ctx, types.ObjectType{
		AttrTypes: getIndexAttrTypes(),
	}, readIndices)
	if diags.HasError() {
		return diags
	}
	model.ReadIndices = readIndicesSet

	return nil
}

// indexFromAlias converts a models.IndexAlias to an indexModel
func indexFromAlias(indexName string, aliasData models.IndexAlias) (indexModel, diag.Diagnostics) {
	index := indexModel{
		Name:          types.StringValue(indexName),
		IsHidden:      types.BoolValue(aliasData.IsHidden),
		IndexRouting:  typeutils.NonEmptyStringValue(aliasData.IndexRouting),
		Routing:       typeutils.NonEmptyStringValue(aliasData.Routing),
		SearchRouting: typeutils.NonEmptyStringValue(aliasData.SearchRouting),
	}

	if aliasData.Filter != nil {
		filterBytes, err := json.Marshal(aliasData.Filter)
		if err != nil {
			return indexModel{}, diag.Diagnostics{
				diag.NewErrorDiagnostic("failed to marshal alias filter", err.Error()),
			}
		}
		index.Filter = jsontypes.NewNormalizedValue(string(filterBytes))
	}

	return index, nil
}

func (model *tfModel) toAliasConfigs(ctx context.Context) ([]AliasIndexConfig, diag.Diagnostics) {
	var configs []AliasIndexConfig

	// Handle write index
	if !model.WriteIndex.IsNull() {
		var writeIndex indexModel
		diags := model.WriteIndex.As(ctx, &writeIndex, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			return nil, diags
		}

		config, configDiags := indexToConfig(writeIndex, true)
		if configDiags.HasError() {
			return nil, configDiags
		}
		configs = append(configs, config)
	}

	// Handle read indices
	if !model.ReadIndices.IsNull() {
		var readIndices []indexModel
		diags := model.ReadIndices.ElementsAs(ctx, &readIndices, false)
		if diags.HasError() {
			return nil, diags
		}

		for _, readIndex := range readIndices {
			config, configDiags := indexToConfig(readIndex, false)
			if configDiags.HasError() {
				return nil, configDiags
			}
			configs = append(configs, config)
		}
	}

	return configs, nil
}

// indexToConfig converts an indexModel to AliasIndexConfig
func indexToConfig(index indexModel, isWriteIndex bool) (AliasIndexConfig, diag.Diagnostics) {
	config := AliasIndexConfig{
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
			return AliasIndexConfig{}, diags
		}
	}

	return config, nil
}
