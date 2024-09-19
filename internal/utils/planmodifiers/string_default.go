package planmodifiers

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func StringUseDefaultIfUnknown(defaultValue string) stringDefault {
	return stringDefault{defaultValue: defaultValue}
}

type stringDefault struct {
	defaultValue string
}

func (bd stringDefault) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// Do nothing if there is a known planned value.
	if !req.PlanValue.IsUnknown() {
		return
	}

	// Do nothing if there is an unknown configuration value, otherwise interpolation gets messed up.
	if req.ConfigValue.IsUnknown() {
		return
	}

	resp.PlanValue = basetypes.NewStringValue(bd.defaultValue)
}

func (bd stringDefault) Description(context.Context) string {
	return fmt.Sprintf("Sets the value to [%s] if unknown", bd.defaultValue)
}

func (bd stringDefault) MarkdownDescription(ctx context.Context) string {
	return bd.Description(ctx)
}
