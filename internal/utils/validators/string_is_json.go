package validators

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

type StringIsJSON struct{}

func (s StringIsJSON) Description(_ context.Context) string {
	return "Ensure that the attribute contains valid JSON"
}

func (s StringIsJSON) MarkdownDescription(ctx context.Context) string {
	return s.Description(ctx)
}

func (s StringIsJSON) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	var jsonData interface{}
	if err := json.Unmarshal([]byte(req.ConfigValue.ValueString()), &jsonData); err != nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"expected value to be valid JSON",
			fmt.Sprintf("The provided value is not valid JSON: %s", err),
		)
		return
	}
}
