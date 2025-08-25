package validators

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

type uuidValidator struct{}

func IsUUID() validator.String {
	return uuidValidator{}
}

func (v uuidValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if !utils.IsKnown(req.ConfigValue) {
		return
	}

	_, err := uuid.ParseUUID(req.ConfigValue.ValueString())
	if err == nil {
		return
	}

	resp.Diagnostics.AddAttributeError(
		req.Path,
		"Invalid UUID",
		fmt.Sprintf("Expected a valid UUID, got %s. Parsing error: %v", req.ConfigValue.ValueString(), err),
	)
}

func (v uuidValidator) Description(_ context.Context) string {
	return "value must be a valid UUID in RFC 4122 format"
}

func (v uuidValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}
