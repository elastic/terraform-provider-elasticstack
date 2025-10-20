package role_mapping

import (
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type RoleMappingData struct {
	Id                      types.String         `tfsdk:"id"`
	ElasticsearchConnection types.List           `tfsdk:"elasticsearch_connection"`
	Name                    types.String         `tfsdk:"name"`
	Enabled                 types.Bool           `tfsdk:"enabled"`
	Rules                   jsontypes.Normalized `tfsdk:"rules"`
	Roles                   types.Set            `tfsdk:"roles"`
	RoleTemplates           jsontypes.Normalized `tfsdk:"role_templates"`
	Metadata                jsontypes.Normalized `tfsdk:"metadata"`
}
