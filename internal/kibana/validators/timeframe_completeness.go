package validators

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ validator.Object = TimeframeCompletenessValidator{}

// TimeframeCompletenessValidator validates that all timeframe attributes
// are set when the timeframe block is present.
type TimeframeCompletenessValidator struct{}

func (v TimeframeCompletenessValidator) Description(ctx context.Context) string {
	return "validates that all timeframe attributes are set when the timeframe block is present"
}

func (v TimeframeCompletenessValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v TimeframeCompletenessValidator) ValidateObject(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	// If the block is null or unknown, nothing to validate
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	// Block is present - check all required attributes
	attrs := req.ConfigValue.Attributes()
	requiredAttrs := []string{"days", "timezone", "hours_start", "hours_end"}

	for _, attrName := range requiredAttrs {
		attr, ok := attrs[attrName]
		if !ok || attr.IsNull() || attr.IsUnknown() {
			resp.Diagnostics.AddAttributeError(
				req.Path.AtName(attrName),
				"Missing Required Timeframe Attribute",
				fmt.Sprintf("The '%s' attribute is required when the timeframe block is present. "+
					"According to the Kibana API specification, all timeframe attributes (days, timezone, hours_start, hours_end) "+
					"must be provided when using a timeframe block.", attrName),
			)
		}
	}
}
