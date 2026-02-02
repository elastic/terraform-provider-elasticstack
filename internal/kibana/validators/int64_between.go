package validators

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// Int64Between validates that an int64 is between min and max (inclusive).
type Int64Between struct {
	Min int64
	Max int64
}

func (v Int64Between) Description(_ context.Context) string {
	return fmt.Sprintf("value must be between %d and %d (inclusive)", v.Min, v.Max)
}

func (v Int64Between) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v Int64Between) ValidateInt64(_ context.Context, req validator.Int64Request, resp *validator.Int64Response) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	val := req.ConfigValue.ValueInt64()
	if val < v.Min || val > v.Max {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			fmt.Sprintf("value must be between %d and %d", v.Min, v.Max),
			fmt.Sprintf("This value must be between %d and %d (inclusive), got: %d.", v.Min, v.Max, val),
		)
		return
	}
}
