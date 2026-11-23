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

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
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
		entitycore.KibanaDataSourceOptions[dataSourceModel]{
			Schema: getDataSourceSchema,
			Read:   readOsquerySavedQueryDataSource,
		},
	)
}

func readOsquerySavedQueryDataSource(ctx context.Context, client *clients.KibanaScopedClient, savedQueryID, spaceID string, config dataSourceModel) (dataSourceModel, bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Datasource schema cannot declare stringdefault.StaticString (no Default on datasource
	// StringAttribute in terraform-plugin-framework); default the default space at read time.
	spaceID = clients.EffectiveSpaceID(spaceID)

	entity, getDiags := kibanaoapi.GetOsquerySavedQuery(ctx, client.GetKibanaOapiClient(), spaceID, savedQueryID)
	diags.Append(getDiags...)
	if diags.HasError() {
		return config, false, diags
	}

	return finishOsquerySavedQueryDataSourceRead(ctx, config, entity, spaceID)
}

func finishOsquerySavedQueryDataSourceRead(
	ctx context.Context,
	config dataSourceModel,
	entity *kibanaoapi.OsquerySavedQueryGetEntity,
	spaceID string,
) (dataSourceModel, bool, diag.Diagnostics) {
	if entity == nil {
		return config, false, nil
	}

	config.SpaceID = types.StringValue(spaceID)
	diags := config.populateFromGetAPI(ctx, entity)
	return config, !diags.HasError(), diags
}
