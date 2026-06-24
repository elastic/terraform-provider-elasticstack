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
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = NewDataSource()
	_ datasource.DataSourceWithConfigure = NewDataSource().(datasource.DataSourceWithConfigure)
)

// NewDataSource is a helper function to simplify the provider implementation.
func NewDataSource() datasource.DataSource {
	return entitycore.NewKibanaDataSource[dataSourceModel](
		entitycore.ComponentKibana,
		"osquery_saved_query",
		getDataSourceSchema,
		readOsquerySavedQueryDataSource,
	)
}

func readOsquerySavedQueryDataSource(ctx context.Context, client *clients.KibanaScopedClient, config dataSourceModel) (dataSourceModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	if !typeutils.IsKnown(config.SavedQueryID) || config.SavedQueryID.ValueString() == "" {
		diags.AddError("Invalid configuration", "saved_query_id must be set.")
		return config, diags
	}

	spaceID := resolveDataSourceSpaceID(config.SpaceID)

	savedQueryID := config.SavedQueryID.ValueString()

	entity, getDiags := kibanaoapi.GetOsquerySavedQuery(ctx, client.GetKibanaOapiClient(), spaceID, savedQueryID)
	diags.Append(getDiags...)
	if diags.HasError() {
		return config, diags
	}

	return finishOsquerySavedQueryDataSourceRead(ctx, config, entity, spaceID)
}

func finishOsquerySavedQueryDataSourceRead(
	ctx context.Context,
	config dataSourceModel,
	entity *kibanaoapi.OsquerySavedQueryGetEntity,
	spaceID string,
) (dataSourceModel, diag.Diagnostics) {
	if entity == nil {
		return config, osquerySavedQueryNotFoundDiagnostic(spaceID, config.SavedQueryID.ValueString())
	}

	config.SpaceID = types.StringValue(spaceID)
	diags := config.populateFromGetAPI(ctx, entity)
	return config, diags
}

func resolveDataSourceSpaceID(spaceID types.String) string {
	if typeutils.IsKnown(spaceID) && spaceID.ValueString() != "" {
		return spaceID.ValueString()
	}
	return clients.DefaultSpaceID
}

func osquerySavedQueryNotFoundDiagnostic(spaceID, savedQueryID string) diag.Diagnostics {
	return diag.Diagnostics{
		diag.NewErrorDiagnostic(
			"Osquery saved query not found",
			fmt.Sprintf("No Osquery saved query with ID %q exists in Kibana space %q.", savedQueryID, spaceID),
		),
	}
}
