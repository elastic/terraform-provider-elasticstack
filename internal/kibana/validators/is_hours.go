package validators

import (
	"context"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var hoursPattern = regexp.MustCompile(`^([0-9]|0[0-9]|1[0-9]|2[0-3]):[0-5][0-9]$`)

// StringIsHours validates that a string is in the format HH:mm (24-hour notation).
type StringIsHours struct{}

func (s StringIsHours) Description(_ context.Context) string {
	return "a valid time in 24-hour notation (HH:mm)"
}

func (s StringIsHours) MarkdownDescription(ctx context.Context) string {
	return s.Description(ctx)
}

func (s StringIsHours) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := req.ConfigValue.ValueString()
	if value == "" {
		return
	}

	if !hoursPattern.MatchString(value) {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid time format",
			"This value must be a valid time in 24-hour notation (HH:mm), for example '09:00' or '23:30'.",
		)
	}
}
