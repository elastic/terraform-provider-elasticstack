package index

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func stringIsJSONObject(i interface{}, s string) (warnings []string, errors []error) {
	iStr, ok := i.(string)
	if !ok {
		errors = append(errors, fmt.Errorf("expected type of %s to be string", s))
		return warnings, errors
	}

	m := map[string]interface{}{}
	if err := json.Unmarshal([]byte(iStr), &m); err != nil {
		errors = append(errors, fmt.Errorf("expected %s to be a JSON object. Check the documentation for the expected format. %w", s, err))
		return
	}

	return
}

type StringIsJSONObject struct{}

func (s StringIsJSONObject) Description(_ context.Context) string {
	return "Ensure that the attribute contains a valid JSON object, and not a simple value"
}

func (s StringIsJSONObject) MarkdownDescription(ctx context.Context) string {
	return s.Description(ctx)
}

func (s StringIsJSONObject) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	m := map[string]interface{}{}
	if err := json.Unmarshal([]byte(req.ConfigValue.ValueString()), &m); err != nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"expected value to be a JSON object",
			fmt.Sprintf("This value must be an object, not a simple type or array. Check the documentation for the expected format. %s", err),
		)
		return
	}
}
