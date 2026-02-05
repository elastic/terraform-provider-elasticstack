package dashboard

import (
	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type filterSimpleModel struct {
	Language types.String `tfsdk:"language"`
	Query    types.String `tfsdk:"query"`
}

func (m *filterSimpleModel) fromAPI(apiQuery kbapi.FilterSimpleSchema) {
	m.Query = types.StringValue(apiQuery.Query)
	m.Language = typeutils.StringishPointerValue(apiQuery.Language)
}

func (m *filterSimpleModel) toAPI() kbapi.FilterSimpleSchema {
	if m == nil {
		return kbapi.FilterSimpleSchema{}
	}

	query := kbapi.FilterSimpleSchema{
		Query: m.Query.ValueString(),
	}
	if utils.IsKnown(m.Language) {
		lang := kbapi.FilterSimpleSchemaLanguage(m.Language.ValueString())
		query.Language = &lang
	}
	return query
}
