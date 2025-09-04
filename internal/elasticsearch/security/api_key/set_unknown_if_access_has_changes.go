package api_key

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// SetUnknownIfAccessHasChanges returns a plan modifier that sets the current attribute to unknown
// if the access attribute has changed between state and config for cross-cluster API keys.
func SetUnknownIfAccessHasChanges() planmodifier.String {
	return setUnknownIfAccessHasChanges{}
}

type setUnknownIfAccessHasChanges struct{}

func (s setUnknownIfAccessHasChanges) Description(ctx context.Context) string {
	return "Sets the attribute value to unknown if the access attribute has changed for cross-cluster API keys"
}

func (s setUnknownIfAccessHasChanges) MarkdownDescription(ctx context.Context) string {
	return s.Description(ctx)
}

func (s setUnknownIfAccessHasChanges) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// Only apply this modifier if we have both state and config
	if req.State.Raw.IsNull() || req.Config.Raw.IsNull() {
		return
	}

	// Get the type attribute to check if this is a cross-cluster API key
	var keyType types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("type"), &keyType)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Only apply to cross-cluster API keys
	if keyType.ValueString() != crossClusterAPIKeyType {
		return
	}

	// Get the access attribute from state and config to check if it has changed
	var stateAccess, configAccess types.Object
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("access"), &stateAccess)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("access"), &configAccess)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If the access attribute has changed between state and config, set the current attribute to Unknown
	if !stateAccess.Equal(configAccess) {
		resp.PlanValue = types.StringUnknown()
	}
}
