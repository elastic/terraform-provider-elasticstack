// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package serverhost

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type serverHostModel struct {
	ID       types.String `tfsdk:"id"`
	HostID   types.String `tfsdk:"host_id"`
	Name     types.String `tfsdk:"name"`
	Hosts    types.List   `tfsdk:"hosts"`
	Default  types.Bool   `tfsdk:"default"`
	SpaceIDs types.Set    `tfsdk:"space_ids"` // > string
}

func (model *serverHostModel) populateFromAPI(ctx context.Context, data *kbapi.ServerHost) (diags diag.Diagnostics) {
	if data == nil {
		return nil
	}

	model.ID = types.StringValue(data.Id)
	model.HostID = types.StringValue(data.Id)
	model.Name = types.StringValue(data.Name)
	model.Hosts = typeutils.SliceToListTypeString(ctx, data.HostUrls, path.Root("hosts"), &diags)
	model.Default = types.BoolPointerValue(data.IsDefault)

	// Note: SpaceIDs is not returned by the API for server hosts, so we preserve it from existing state.
	// It's only used to determine which API endpoint to call.
	// If space_ids is unknown (not provided by user), set to null to satisfy Terraform's requirement.
	if model.SpaceIDs.IsUnknown() {
		model.SpaceIDs = types.SetNull(types.StringType)
	}

	return
}

func (model serverHostModel) toAPICreateModel(ctx context.Context) (body kbapi.PostFleetFleetServerHostsJSONRequestBody, diags diag.Diagnostics) {
	body = kbapi.PostFleetFleetServerHostsJSONRequestBody{
		HostUrls:  typeutils.ListTypeToSliceString(ctx, model.Hosts, path.Root("hosts"), &diags),
		Id:        model.HostID.ValueStringPointer(),
		IsDefault: model.Default.ValueBoolPointer(),
		Name:      model.Name.ValueString(),
	}
	return
}

func (model serverHostModel) toAPIUpdateModel(ctx context.Context) (body kbapi.PutFleetFleetServerHostsItemidJSONRequestBody, diags diag.Diagnostics) {
	body = kbapi.PutFleetFleetServerHostsItemidJSONRequestBody{
		HostUrls:  schemautil.SliceRef(typeutils.ListTypeToSliceString(ctx, model.Hosts, path.Root("hosts"), &diags)),
		IsDefault: model.Default.ValueBoolPointer(),
		Name:      model.Name.ValueStringPointer(),
	}
	return
}
