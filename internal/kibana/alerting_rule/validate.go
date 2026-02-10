package alerting_rule

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *Resource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data alertingRuleModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !utils.IsKnown(data.Params) || !utils.IsKnown(data.RuleTypeID) {
		return
	}

	var params map[string]interface{}
	if err := json.Unmarshal([]byte(data.Params.ValueString()), &params); err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("params"),
			"Invalid params JSON",
			err.Error(),
		)
		return
	}

	errs := validateRuleParams(data.RuleTypeID.ValueString(), params)
	if len(errs) == 0 {
		return
	}

	resp.Diagnostics.AddAttributeError(
		path.Root("params"),
		"Invalid params for rule_type_id",
		strings.Join(errs, "; "),
	)
}
