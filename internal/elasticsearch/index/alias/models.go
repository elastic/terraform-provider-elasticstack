package alias

import (
	"context"
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type tfModel struct {
	ID            types.String         `tfsdk:"id"`
	Name          types.String         `tfsdk:"name"`
	Indices       types.Set            `tfsdk:"indices"`
	Filter        jsontypes.Normalized `tfsdk:"filter"`
	IndexRouting  types.String         `tfsdk:"index_routing"`
	IsHidden      types.Bool           `tfsdk:"is_hidden"`
	IsWriteIndex  types.Bool           `tfsdk:"is_write_index"`
	Routing       types.String         `tfsdk:"routing"`
	SearchRouting types.String         `tfsdk:"search_routing"`
}

func (model *tfModel) populateFromAPI(ctx context.Context, aliasName string, aliasData models.IndexAlias, indices []string) diag.Diagnostics {
	model.ID = types.StringValue(aliasName)
	model.Name = types.StringValue(aliasName)

	indicesSet, diags := types.SetValueFrom(ctx, types.StringType, indices)
	if diags.HasError() {
		return diags
	}
	model.Indices = indicesSet

	model.IndexRouting = types.StringValue(aliasData.IndexRouting)
	model.IsHidden = types.BoolValue(aliasData.IsHidden)
	model.IsWriteIndex = types.BoolValue(aliasData.IsWriteIndex)
	model.Routing = types.StringValue(aliasData.Routing)
	model.SearchRouting = types.StringValue(aliasData.SearchRouting)

	if aliasData.Filter != nil {
		filterBytes, err := json.Marshal(aliasData.Filter)
		if err != nil {
			return diag.Diagnostics{
				diag.NewErrorDiagnostic("failed to marshal alias filter", err.Error()),
			}
		}
		model.Filter = jsontypes.NewNormalizedValue(string(filterBytes))
	}

	return nil
}

func (model *tfModel) toAPIModel() (models.IndexAlias, []string, diag.Diagnostics) {
	apiModel := models.IndexAlias{
		Name:          model.Name.ValueString(),
		IndexRouting:  model.IndexRouting.ValueString(),
		IsHidden:      model.IsHidden.ValueBool(),
		IsWriteIndex:  model.IsWriteIndex.ValueBool(),
		Routing:       model.Routing.ValueString(),
		SearchRouting: model.SearchRouting.ValueString(),
	}

	if utils.IsKnown(model.Filter) {
		if diags := model.Filter.Unmarshal(&apiModel.Filter); diags.HasError() {
			return models.IndexAlias{}, nil, diags
		}
	}

	var indices []string
	diags := model.Indices.ElementsAs(context.Background(), &indices, false)
	if diags.HasError() {
		return models.IndexAlias{}, nil, diags
	}

	return apiModel, indices, nil
}