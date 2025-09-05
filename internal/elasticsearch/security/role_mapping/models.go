package role_mapping

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type RoleMappingData struct {
	Id                      types.String `tfsdk:"id"`
	ElasticsearchConnection types.List   `tfsdk:"elasticsearch_connection"`
	Name                    types.String `tfsdk:"name"`
	Enabled                 types.Bool   `tfsdk:"enabled"`
	Rules                   types.String `tfsdk:"rules"`
	Roles                   types.Set    `tfsdk:"roles"`
	RoleTemplates           types.String `tfsdk:"role_templates"`
	Metadata                types.String `tfsdk:"metadata"`
}
