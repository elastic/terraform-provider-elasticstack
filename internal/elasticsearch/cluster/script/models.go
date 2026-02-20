package script

import (
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type Data struct {
	ID                      types.String         `tfsdk:"id"`
	ElasticsearchConnection types.List           `tfsdk:"elasticsearch_connection"`
	ScriptID                types.String         `tfsdk:"script_id"`
	Lang                    types.String         `tfsdk:"lang"`
	Source                  types.String         `tfsdk:"source"`
	Params                  jsontypes.Normalized `tfsdk:"params"`
	Context                 types.String         `tfsdk:"context"`
}
