package server_host

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type serverHostModel struct {
	Id       types.String `tfsdk:"id"`
	HostID   types.String `tfsdk:"host_id"`
	Name     types.String `tfsdk:"name"`
	Hosts    types.List   `tfsdk:"hosts"`
	Default  types.Bool   `tfsdk:"default"`
	SpaceIds types.List   `tfsdk:"space_ids"` //> string
}

func (model *serverHostModel) populateFromAPI(ctx context.Context, data *kbapi.ServerHost) (diags diag.Diagnostics) {
	if data == nil {
		return nil
	}

	model.Id = types.StringValue(data.Id)
	model.HostID = types.StringValue(data.Id)
	model.Name = types.StringValue(data.Name)
	model.Hosts = utils.SliceToListType_String(ctx, data.HostUrls, path.Root("hosts"), &diags)
	model.Default = types.BoolPointerValue(data.IsDefault)

	// Note: SpaceIds is not returned by the API for server hosts, so we preserve it from existing state
	// It's only used to determine which API endpoint to call
	if model.SpaceIds.IsNull() {
		model.SpaceIds = types.ListNull(types.StringType)
	}

	return
}

func (model serverHostModel) toAPICreateModel(ctx context.Context) (body kbapi.PostFleetFleetServerHostsJSONRequestBody, diags diag.Diagnostics) {
	body = kbapi.PostFleetFleetServerHostsJSONRequestBody{
		HostUrls:  utils.ListTypeToSlice_String(ctx, model.Hosts, path.Root("hosts"), &diags),
		Id:        model.HostID.ValueStringPointer(),
		IsDefault: model.Default.ValueBoolPointer(),
		Name:      model.Name.ValueString(),
	}
	return
}

func (model serverHostModel) toAPIUpdateModel(ctx context.Context) (body kbapi.PutFleetFleetServerHostsItemidJSONRequestBody, diags diag.Diagnostics) {
	body = kbapi.PutFleetFleetServerHostsItemidJSONRequestBody{
		HostUrls:  utils.SliceRef(utils.ListTypeToSlice_String(ctx, model.Hosts, path.Root("hosts"), &diags)),
		IsDefault: model.Default.ValueBoolPointer(),
		Name:      model.Name.ValueStringPointer(),
	}
	return
}
