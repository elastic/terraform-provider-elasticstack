package validators

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func StringMatchesISO8601Regex(s string) (matched bool, err error) {
	pattern := `(\d{4}-[01]\d-[0-3]\dT[0-2]\d:[0-5]\d:[0-5]\d\.\d+([+-][0-2]\d:[0-5]\d|Z))|(\d{4}-[01]\d-[0-3]\dT[0-2]\d:[0-5]\d:[0-5]\d([+-][0-2]\d:[0-5]\d|Z))|(\d{4}-[01]\d-[0-3]\dT[0-2]\d:[0-5]\d([+-][0-2]\d:[0-5]\d|Z))`
	return regexp.MatchString(pattern, s)
}

type StringIsISO8601 struct{}

func (s StringIsISO8601) Description(_ context.Context) string {
	return "a valid ISO8601 date and time formatted string"
}

func (s StringIsISO8601) MarkdownDescription(ctx context.Context) string {
	return s.Description(ctx)
}

func (s StringIsISO8601) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	if matched, err := StringMatchesISO8601Regex(req.ConfigValue.ValueString()); err != nil || !matched {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"expected value to be a valid ISO8601 string",
			fmt.Sprintf("This value must be a valid ISO8601 date and time formatted string %s", err),
		)
		return
	}
}
