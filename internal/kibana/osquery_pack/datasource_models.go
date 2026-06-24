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

package osquerypack

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type dataSourceModel struct {
	entitycore.KibanaConnectionField
	ID          types.String `tfsdk:"id"`
	PackID      types.String `tfsdk:"pack_id"`
	SpaceID     types.String `tfsdk:"space_id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	PolicyIDs   types.List   `tfsdk:"policy_ids"`
	Shards      types.Map    `tfsdk:"shards"`
	Queries     types.Map    `tfsdk:"queries"`
	ReadOnly    types.Bool   `tfsdk:"read_only"`
}

func (dataSourceModel) GetVersionRequirements(_ context.Context) ([]entitycore.VersionRequirement, diag.Diagnostics) {
	return []entitycore.VersionRequirement{
		{
			MinVersion:   *osqueryPackMinVersion,
			ErrorMessage: fmt.Sprintf("Osquery packs require Elastic Stack v%s or later.", osqueryPackMinVersion),
		},
	}, nil
}

func (m *dataSourceModel) populateFromAPI(ctx context.Context, spaceID string, data *kibanaoapi.OsqueryPackDetail) diag.Diagnostics {
	var pack osqueryPackModel
	diags := pack.populateFromAPI(ctx, spaceID, data)
	if diags.HasError() {
		return diags
	}

	m.ID = pack.ID
	m.PackID = pack.PackID
	m.SpaceID = pack.SpaceID
	m.Name = pack.Name
	m.Description = pack.Description
	m.Enabled = pack.Enabled
	m.PolicyIDs = pack.PolicyIDs
	m.Shards = pack.Shards
	m.Queries = pack.Queries
	m.ReadOnly = types.BoolPointerValue(data.ReadOnly)

	return diags
}
