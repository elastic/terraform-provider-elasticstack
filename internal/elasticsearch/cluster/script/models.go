package script

import (
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ScriptData struct {
	Id                      types.String         `tfsdk:"id"`
	ElasticsearchConnection types.List           `tfsdk:"elasticsearch_connection"`
	ScriptId                types.String         `tfsdk:"script_id"`
	Lang                    types.String         `tfsdk:"lang"`
	Source                  types.String         `tfsdk:"source"`
	Params                  jsontypes.Normalized `tfsdk:"params"`
	Context                 types.String         `tfsdk:"context"`
}
