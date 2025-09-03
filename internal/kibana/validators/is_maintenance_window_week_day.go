package validators

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

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
