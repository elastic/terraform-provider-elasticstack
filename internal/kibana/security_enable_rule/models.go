package security_enable_rule

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type enableRuleModel struct {
	ID               types.String `tfsdk:"id"`
	SpaceID          types.String `tfsdk:"space_id"`
	Key              types.String `tfsdk:"key"`
	Value            types.String `tfsdk:"value"`
	DisableOnDestroy types.Bool   `tfsdk:"disable_on_destroy"`
	AllRulesEnabled  types.Bool   `tfsdk:"all_rules_enabled"`
}
