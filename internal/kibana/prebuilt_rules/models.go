package prebuilt_rules

import (
	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type prebuiltRuleModel struct {
	ID                    types.String `tfsdk:"id"`
	SpaceID               types.String `tfsdk:"space_id"`
	RulesInstalled        types.Int64  `tfsdk:"rules_installed"`
	RulesNotInstalled     types.Int64  `tfsdk:"rules_not_installed"`
	RulesNotUpdated       types.Int64  `tfsdk:"rules_not_updated"`
	TimelinesInstalled    types.Int64  `tfsdk:"timelines_installed"`
	TimelinesNotInstalled types.Int64  `tfsdk:"timelines_not_installed"`
	TimelinesNotUpdated   types.Int64  `tfsdk:"timelines_not_updated"`
}

func (model *prebuiltRuleModel) populateFromStatus(status *kbapi.ReadPrebuiltRulesAndTimelinesStatusResponse) {
	model.RulesInstalled = types.Int64Value(int64(status.JSON200.RulesInstalled))
	model.RulesNotInstalled = types.Int64Value(int64(status.JSON200.RulesNotInstalled))
	model.RulesNotUpdated = types.Int64Value(int64(status.JSON200.RulesNotUpdated))
	model.TimelinesInstalled = types.Int64Value(int64(status.JSON200.TimelinesInstalled))
	model.TimelinesNotInstalled = types.Int64Value(int64(status.JSON200.TimelinesNotInstalled))
	model.TimelinesNotUpdated = types.Int64Value(int64(status.JSON200.TimelinesNotUpdated))
}
