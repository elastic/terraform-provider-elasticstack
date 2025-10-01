package alias

import (
	"context"
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
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

type writeIndexModel struct {
	Name          types.String         `tfsdk:"name"`
	Filter        jsontypes.Normalized `tfsdk:"filter"`
	IndexRouting  types.String         `tfsdk:"index_routing"`
	IsHidden      types.Bool           `tfsdk:"is_hidden"`
	Routing       types.String         `tfsdk:"routing"`
	SearchRouting types.String         `tfsdk:"search_routing"`
}

type readIndexModel struct {
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

func (model *tfModel) populateFromAPI(ctx context.Context, aliasName string, indices map[string]models.IndexAlias) diag.Diagnostics {
	model.Name = types.StringValue(aliasName)

	var writeIndex *writeIndexModel
	var readIndices []readIndexModel

	for indexName, aliasData := range indices {
		if aliasData.IsWriteIndex {
			writeIndex = &writeIndexModel{
				Name:     types.StringValue(indexName),
				IsHidden: types.BoolValue(aliasData.IsHidden),
			}

			if aliasData.IndexRouting != "" {
				writeIndex.IndexRouting = types.StringValue(aliasData.IndexRouting)
			}
			if aliasData.Routing != "" {
				writeIndex.Routing = types.StringValue(aliasData.Routing)
			}
			if aliasData.SearchRouting != "" {
				writeIndex.SearchRouting = types.StringValue(aliasData.SearchRouting)
			}
			if aliasData.Filter != nil {
				filterBytes, err := json.Marshal(aliasData.Filter)
				if err != nil {
					return diag.Diagnostics{
						diag.NewErrorDiagnostic("failed to marshal alias filter", err.Error()),
					}
				}
				writeIndex.Filter = jsontypes.NewNormalizedValue(string(filterBytes))
			}
		} else {
			readIndex := readIndexModel{
				Name:     types.StringValue(indexName),
				IsHidden: types.BoolValue(aliasData.IsHidden),
			}

			if aliasData.IndexRouting != "" {
				readIndex.IndexRouting = types.StringValue(aliasData.IndexRouting)
			}
			if aliasData.Routing != "" {
				readIndex.Routing = types.StringValue(aliasData.Routing)
			}
			if aliasData.SearchRouting != "" {
				readIndex.SearchRouting = types.StringValue(aliasData.SearchRouting)
			}
			if aliasData.Filter != nil {
				filterBytes, err := json.Marshal(aliasData.Filter)
				if err != nil {
					return diag.Diagnostics{
						diag.NewErrorDiagnostic("failed to marshal alias filter", err.Error()),
					}
				}
				readIndex.Filter = jsontypes.NewNormalizedValue(string(filterBytes))
			}

			readIndices = append(readIndices, readIndex)
		}
	}

	// Set write index
	if writeIndex != nil {
		writeIndexObj, diags := types.ObjectValueFrom(ctx, writeIndexModel{}.attrTypes(), *writeIndex)
		if diags.HasError() {
			return diags
		}
		model.WriteIndex = writeIndexObj
	} else {
		model.WriteIndex = types.ObjectNull(writeIndexModel{}.attrTypes())
	}

	// Set read indices
	readIndicesSet, diags := types.SetValueFrom(ctx, types.ObjectType{
		AttrTypes: readIndexModel{}.attrTypes(),
	}, readIndices)
	if diags.HasError() {
		return diags
	}
	model.ReadIndices = readIndicesSet

	return nil
}

func (model *tfModel) toAliasConfigs(ctx context.Context) ([]AliasIndexConfig, diag.Diagnostics) {
	var configs []AliasIndexConfig

	// Handle write index
	if !model.WriteIndex.IsNull() {
		var writeIndex writeIndexModel
		diags := model.WriteIndex.As(ctx, &writeIndex, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			return nil, diags
		}

		config := AliasIndexConfig{
			Name:         writeIndex.Name.ValueString(),
			IsWriteIndex: true,
			IsHidden:     writeIndex.IsHidden.ValueBool(),
		}

		if !writeIndex.IndexRouting.IsNull() {
			config.IndexRouting = writeIndex.IndexRouting.ValueString()
		}
		if !writeIndex.Routing.IsNull() {
			config.Routing = writeIndex.Routing.ValueString()
		}
		if !writeIndex.SearchRouting.IsNull() {
			config.SearchRouting = writeIndex.SearchRouting.ValueString()
		}
		if !writeIndex.Filter.IsNull() {
			if diags := writeIndex.Filter.Unmarshal(&config.Filter); diags.HasError() {
				return nil, diags
			}
		}

		configs = append(configs, config)
	}

	// Handle read indices
	if !model.ReadIndices.IsNull() {
		var readIndices []readIndexModel
		diags := model.ReadIndices.ElementsAs(ctx, &readIndices, false)
		if diags.HasError() {
			return nil, diags
		}

		for _, readIndex := range readIndices {
			config := AliasIndexConfig{
				Name:         readIndex.Name.ValueString(),
				IsWriteIndex: false,
				IsHidden:     readIndex.IsHidden.ValueBool(),
			}

			if !readIndex.IndexRouting.IsNull() {
				config.IndexRouting = readIndex.IndexRouting.ValueString()
			}
			if !readIndex.Routing.IsNull() {
				config.Routing = readIndex.Routing.ValueString()
			}
			if !readIndex.SearchRouting.IsNull() {
				config.SearchRouting = readIndex.SearchRouting.ValueString()
			}
			if !readIndex.Filter.IsNull() {
				if diags := readIndex.Filter.Unmarshal(&config.Filter); diags.HasError() {
					return nil, diags
				}
			}

			configs = append(configs, config)
		}
	}

	return configs, nil
}

// Helper functions for attribute types
func (writeIndexModel) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":           types.StringType,
		"filter":         jsontypes.NormalizedType{},
		"index_routing":  types.StringType,
		"is_hidden":      types.BoolType,
		"routing":        types.StringType,
		"search_routing": types.StringType,
	}
}

func (readIndexModel) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":           types.StringType,
		"filter":         jsontypes.NormalizedType{},
		"index_routing":  types.StringType,
		"is_hidden":      types.BoolType,
		"routing":        types.StringType,
		"search_routing": types.StringType,
	}
}