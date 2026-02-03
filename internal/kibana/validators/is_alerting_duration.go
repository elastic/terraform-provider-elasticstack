package validators

import (
	"context"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var alertingDurationPattern = "^[1-9][0-9]*(?:d|h|m|s)$"

func StringMatchesAlertingDurationRegex(s string) (matched bool, err error) {
	return regexp.MatchString(alertingDurationPattern, s)
}

type StringIsAlertingDuration struct{}

func (s StringIsAlertingDuration) Description(_ context.Context) string {
	return "a valid alerting duration in seconds (s), minutes (m), hours (h), or days (d)"
}

func (s StringIsAlertingDuration) MarkdownDescription(ctx context.Context) string {
	return s.Description(ctx)
}

func (s StringIsAlertingDuration) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	if matched, err := StringMatchesAlertingDurationRegex(req.ConfigValue.ValueString()); err != nil || !matched {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"expected value to be a valid alerting duration",
			"This value must be a valid alerting duration in seconds (s), minutes (m), hours (h), or days (d).",
		)
		return
	}
}
