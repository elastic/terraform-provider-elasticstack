package dashboard

import (
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type searchFilterModel struct {
	Query    types.String         `tfsdk:"query"`
	MetaJSON jsontypes.Normalized `tfsdk:"meta_json"`
	Language types.String         `tfsdk:"language"`
}

func (m *searchFilterModel) fromAPI(apiFilter kbapi.SearchFilterSchema) diag.Diagnostics {
	var diags diag.Diagnostics

	// Try to extract from SearchFilterSchema0
	filterSchema, err := apiFilter.AsSearchFilterSchema0()
	if err != nil {
		diags.AddError("Failed to extract search filter", err.Error())
		return diags
	}

	// Extract string from union type
	queryStr, queryErr := filterSchema.Query.AsSearchFilterSchema0Query0()
	if queryErr != nil {
		diags.AddError("Failed to extract search filter query", queryErr.Error())
		return diags
	}

	m.Query = types.StringValue(queryStr)

	// Language defaults to "kuery" if the API doesn't return it
	// This is consistent with Kibana's default behavior
	if filterSchema.Language != nil {
		m.Language = types.StringValue(string(*filterSchema.Language))
	} else {
		m.Language = types.StringValue("kuery")
	}

	if filterSchema.Meta != nil {
		metaJSON, err := json.Marshal(filterSchema.Meta)
		if err == nil {
			m.MetaJSON = jsontypes.NewNormalizedValue(string(metaJSON))
		}
	}

	return diags
}

func (m *searchFilterModel) toAPI() (kbapi.SearchFilterSchema, diag.Diagnostics) {
	var diags diag.Diagnostics

	filter := kbapi.SearchFilterSchema0{}
	if utils.IsKnown(m.Query) {
		query := m.Query.ValueString()
		var queryUnion kbapi.SearchFilterSchema_0_Query
		if err := queryUnion.FromSearchFilterSchema0Query0(query); err != nil {
			diags.AddError("Failed to create search filter query", err.Error())
			return kbapi.SearchFilterSchema{}, diags
		}
		filter.Query = queryUnion
	}
	if utils.IsKnown(m.Language) {
		lang := kbapi.SearchFilterSchema0Language(m.Language.ValueString())
		filter.Language = &lang
	}
	if utils.IsKnown(m.MetaJSON) {
		var meta map[string]interface{}
		diags.Append(m.MetaJSON.Unmarshal(&meta)...)
		if !diags.HasError() {
			filter.Meta = &meta
		}
	}

	var result kbapi.SearchFilterSchema
	if err := result.FromSearchFilterSchema0(filter); err != nil {
		diags.AddError("Failed to create search filter", err.Error())
	}
	return result, diags
}
