package datafeed_state

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// SetUnknownIfStateHasChanges returns a plan modifier that sets the current attribute to unknown
// if the state attribute has changed between state and config.
func SetUnknownIfStateHasChanges() planmodifier.String {
	return setUnknownIfStateHasChanges{}
}

type setUnknownIfStateHasChanges struct{}

func (s setUnknownIfStateHasChanges) Description(ctx context.Context) string {
	return "Sets the attribute value to unknown if the state attribute has changed"
}

func (s setUnknownIfStateHasChanges) MarkdownDescription(ctx context.Context) string {
	return s.Description(ctx)
}

func (s setUnknownIfStateHasChanges) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// Only apply this modifier if we have both state and config
	if req.State.Raw.IsNull() || req.Config.Raw.IsNull() {
		return
	}

	// Continue using the config value if it's explicitly set
	if utils.IsKnown(req.ConfigValue) {
		return
	}

	// Get the state attribute from state and config to check if it has changed
	var stateValue, configValue types.String
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("state"), &stateValue)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("state"), &configValue)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If the state attribute has changed between state and config, set the current attribute to Unknown
	if !stateValue.Equal(configValue) {
		resp.PlanValue = types.StringUnknown()
	}
}
