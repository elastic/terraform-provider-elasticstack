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

package osquerysavedquery

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ entitycore.WithVersionRequirements = dataSourceModel{}

type dataSourceModel struct {
	entitycore.KibanaConnectionField

	ID            types.String `tfsdk:"id"`
	SavedObjectID types.String `tfsdk:"saved_object_id"`
	SavedQueryID  types.String `tfsdk:"saved_query_id"`
	SpaceID       types.String `tfsdk:"space_id"`
	Query         types.String `tfsdk:"query"`
	Description   types.String `tfsdk:"description"`
	Platform      types.Set    `tfsdk:"platform"`
	Interval      types.Int64  `tfsdk:"interval"`
	Version       types.String `tfsdk:"version"`
	Snapshot      types.Bool   `tfsdk:"snapshot"`
	Removed       types.Bool   `tfsdk:"removed"`
	EcsMapping    types.Map    `tfsdk:"ecs_mapping"`
	Prebuilt      types.Bool   `tfsdk:"prebuilt"`
}

func (m dataSourceModel) GetVersionRequirements(ctx context.Context) ([]entitycore.VersionRequirement, diag.Diagnostics) {
	return osquerySavedQueryModel{}.GetVersionRequirements(ctx)
}

func (m *dataSourceModel) populateFromGetAPI(ctx context.Context, entity *kibanaoapi.OsquerySavedQueryGetEntity) diag.Diagnostics {
	if entity == nil {
		return nil
	}

	scratch := osquerySavedQueryModel{
		SavedQueryID: m.SavedQueryID,
		SpaceID:      m.SpaceID,
	}
	diags := scratch.populateFromGetAPI(ctx, entity)
	if diags.HasError() {
		return diags
	}

	m.ID = scratch.ID
	m.SavedObjectID = scratch.SavedObjectID
	m.SavedQueryID = scratch.SavedQueryID
	m.SpaceID = scratch.SpaceID
	m.Query = scratch.Query
	m.Description = scratch.Description
	m.Platform = scratch.Platform
	m.Interval = scratch.Interval
	m.Version = scratch.Version
	m.Snapshot = scratch.Snapshot
	m.Removed = scratch.Removed
	m.EcsMapping = scratch.EcsMapping
	m.Prebuilt = prebuiltFromAPI(entity.Prebuilt)

	return diags
}

// prebuiltFromAPI maps the API prebuilt flag to state. Omitted/nil is treated as false
// so user-managed queries surface prebuilt = false rather than null.
func prebuiltFromAPI(prebuilt *bool) types.Bool {
	if prebuilt == nil {
		return types.BoolValue(false)
	}

	return types.BoolValue(*prebuilt)
}
