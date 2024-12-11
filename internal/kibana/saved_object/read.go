package saved_object

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model ksoModelV0

	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	kibanaClient, err := r.client.GetKibanaClient()
	if err != nil {
		resp.Diagnostics.AddError("unable to get kibana client", err.Error())
		return
	}

	result, err := kibanaClient.KibanaSavedObject.Get(model.Type.ValueString(), model.ID.ValueString(), model.SpaceID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("failed to get saved object", err.Error())
		return
	}

	ksoRemoveUnwantedFields(result)

	object, err := json.Marshal(result)
	if err != nil {
		resp.Diagnostics.AddError("failed to marshal saved object", err.Error())
		return
	}

	model.Imported = types.StringValue(string(object))

	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
	if resp.Diagnostics.HasError() {
		return
	}
}
