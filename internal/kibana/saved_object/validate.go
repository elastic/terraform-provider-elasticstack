package saved_object

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// KibanaSavedObjectValidator is the underlying struct implementing ConflictsWith.
type KibanaSavedObjectValidator struct {
}

func (v KibanaSavedObjectValidator) Description(ctx context.Context) string {
	return v.MarkdownDescription(ctx)
}

func (v KibanaSavedObjectValidator) MarkdownDescription(_ context.Context) string {
	return "The Kibana object is must have a 'type' and 'id' field"
}

func (v KibanaSavedObjectValidator) ValidateResource(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var configData ksoModelV0

	resp.Diagnostics.Append(req.Config.Get(ctx, &configData)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var object map[string]any
	err := json.Unmarshal([]byte(configData.Object.ValueString()), &object)
	if err != nil {
		resp.Diagnostics.AddError("invalid JSON in object", err.Error())
		return
	}

	if _, ok := object["type"]; !ok {
		resp.Diagnostics.AddError("missing 'type' field in JSON object", "")
	}
	if _, ok := object["id"]; !ok {
		resp.Diagnostics.AddError("missing 'id' field in JSON object", "")
	}
}
