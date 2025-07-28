package enrich

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type EnrichPolicyData struct {
	Id                      types.String `tfsdk:"id"`
	ElasticsearchConnection types.List   `tfsdk:"elasticsearch_connection"`
	Name                    types.String `tfsdk:"name"`
	PolicyType              types.String `tfsdk:"policy_type"`
	Indices                 types.List   `tfsdk:"indices"`
	MatchField              types.String `tfsdk:"match_field"`
	EnrichFields            types.List   `tfsdk:"enrich_fields"`
	Query                   types.String `tfsdk:"query"`
	Execute                 types.Bool   `tfsdk:"execute"`
}