package validators

import (
	"context"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// StringMatchesHoursRegex checks if the string matches HH:mm format.
func StringMatchesHoursRegex(s string) (matched bool, err error) {
	pattern := `^([0-1]?[0-9]|2[0-3]):[0-5][0-9]$`
	return regexp.MatchString(pattern, s)
}

// StringIsHours validates that a string is in HH:mm format.
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

	if matched, err := StringMatchesHoursRegex(req.ConfigValue.ValueString()); err != nil || !matched {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"expected value to be a valid time in 24-hour notation (HH:mm)",
			"This value must be a valid time in 24-hour notation (HH:mm). For example: 09:00, 14:30, 23:59.",
		)
		return
	}
}
