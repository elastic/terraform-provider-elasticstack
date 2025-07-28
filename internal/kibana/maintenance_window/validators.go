package maintenance_window

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

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

	pattern := `^[1-9][0-9]*(?:d|w|M|y)$`
	matched, err := regexp.MatchString(pattern, req.ConfigValue.ValueString())

	if err != nil || !matched {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"expected value to be a valid interval/frequency",
			fmt.Sprintf("This value must be a valid interval/frequency. Allowed values are in the `<integer><unit>` format. `<unit>` is one of `d`, `w`, `M`, or `y` for days, weeks, months, years. For example: `15d`, `2w`, `3m`, `1y`. %s", err),
		)
		return
	}
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

	pattern := `^(((\+|-)[1-5])?(MO|TU|WE|TH|FR|SA|SU))$`

	if matched, err := regexp.MatchString(pattern, req.ConfigValue.ValueString()); err != nil || !matched {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"expected value to be a valid OnWeekDay",
			fmt.Sprintf("This value must be a valid OnWeekDay. Accepted values are specific days of the week (`[MO,TU,WE,TH,FR,SA,SU]`) or nth day of month (`[+1MO, -3FR, +2WE, -4SA, -5SU]`). %s", err),
		)
		return
	}
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

	pattern := "^[1-9][0-9]*(?:d|h|m|s)$"

	if matched, err := regexp.MatchString(pattern, req.ConfigValue.ValueString()); err != nil || !matched {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"expected value to be a valid alerting duration",
			fmt.Sprintf("This value must be a valid alerting duration in seconds (s), minutes (m), hours (h), or days (d) %s", err),
		)
		return
	}
}
