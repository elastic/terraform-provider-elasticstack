package validation_utils

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
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
			fmt.Sprintf("This value must be a valid interval/frequency. Allowed values are in the `<integer><unit>` format. `<unit>` is one of `d`, `w`, `M`, or `y` for days, weeks, months, years. For example: `15d`, `2w`, `3m`, `1y`. %s", err),
		)
		return
	}
}

func StringMatchesOnWeekDayRegex(s string) (matched bool, err error) {
	pattern := `^(((\+|-)[1-5])?(MO|TU|WE|TH|FR|SA|SU))$`
	return regexp.MatchString(pattern, s)
}

type StringIsMaintenanceWindowOnWeekDay struct{}

func (s StringIsMaintenanceWindowOnWeekDay) Description(_ context.Context) string {
	return "a valid OnWeekDay. Accepted values are specific days of the week (`[MO,TU,WE,TH,FR,SA,SU]`) or nth day of month (`[+1MO, -3FR, +2WE, -4SA, -5SU]`)."
}

func (s StringIsMaintenanceWindowOnWeekDay) MarkdownDescription(ctx context.Context) string {
	return s.Description(ctx)
}

func (s StringIsMaintenanceWindowOnWeekDay) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	if matched, err := StringMatchesOnWeekDayRegex(req.ConfigValue.ValueString()); err != nil || !matched {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"expected value to be a valid OnWeekDay",
			fmt.Sprintf("This value must be a valid OnWeekDay. Accepted values are specific days of the week (`[MO,TU,WE,TH,FR,SA,SU]`) or nth day of month (`[+1MO, -3FR, +2WE, -4SA, -5SU]`). %s", err),
		)
		return
	}
}

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
			fmt.Sprintf("This value must be a valid alerting duration in seconds (s), minutes (m), hours (h), or days (d) %s", err),
		)
		return
	}
}

// Avoid lint error on deprecated SchemaValidateFunc usage.
//
//nolint:staticcheck
func StringIsAlertingDurationSDKV2() schema.SchemaValidateFunc {
	r := regexp.MustCompile(alertingDurationPattern)
	return validation.StringMatch(r, "string is not a valid Alerting duration in seconds (s), minutes (m), hours (h), or days (d)")
}
