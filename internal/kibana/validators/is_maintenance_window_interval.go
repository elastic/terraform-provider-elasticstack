package validators

import (
	"context"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func StringMatchesIntervalFrequencyRegex(s string) (matched bool, err error) {
	pattern := `^[1-9][0-9]*(?:d|w|M|y)$`
	return regexp.MatchString(pattern, s)
}

type StringIsMaintenanceWindowIntervalFrequency struct{}

func (s StringIsMaintenanceWindowIntervalFrequency) Description(_ context.Context) string {
	return "a valid interval/frequency. Allowed values are in the `<integer><unit>` format. `<unit>` is one of `d`, `w`, `M`, or `y` for days, weeks, months, years. For example: `15d`, `2w`, `3m`, `1y`."
}

func (s StringIsMaintenanceWindowIntervalFrequency) MarkdownDescription(ctx context.Context) string {
	return s.Description(ctx)
}

func (s StringIsMaintenanceWindowIntervalFrequency) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	if matched, err := StringMatchesIntervalFrequencyRegex(req.ConfigValue.ValueString()); err != nil || !matched {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"expected value to be a valid interval/frequency",
			"This value must be a valid interval/frequency. Allowed values are in the `<integer><unit>` format. `<unit>` is one of `d`, `w`, `M`, or `y` for days, weeks, months, years. For example: `15d`, `2w`, `3m`, `1y`.",
		)
		return
	}
}
