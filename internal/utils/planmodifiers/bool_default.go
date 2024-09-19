package planmodifiers

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func BoolUseDefaultIfUnknown(defaultValue bool) boolDefault {
	return boolDefault{defaultValue: defaultValue}
}

type boolDefault struct {
	defaultValue bool
}

func (bd boolDefault) PlanModifyBool(ctx context.Context, req planmodifier.BoolRequest, resp *planmodifier.BoolResponse) {
	// Do nothing if there is a known planned value.
	if !req.PlanValue.IsUnknown() {
		return
	}

	// Do nothing if there is an unknown configuration value, otherwise interpolation gets messed up.
	if req.ConfigValue.IsUnknown() {
		return
	}

	resp.PlanValue = basetypes.NewBoolValue(bd.defaultValue)
}

func (bd boolDefault) Description(context.Context) string {
	return fmt.Sprintf("Sets the value to [%t] if unknown", bd.defaultValue)
}

func (bd boolDefault) MarkdownDescription(ctx context.Context) string {
	return bd.Description(ctx)
}
