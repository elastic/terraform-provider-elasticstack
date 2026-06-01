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

package entities

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

var (
	_ datasource.DataSourceWithValidateConfig = &entitiesDataSource{}
)

type entitiesDataSource struct {
	inner datasource.DataSource
}

func NewDataSource() datasource.DataSource {
	return &entitiesDataSource{
		inner: entitycore.NewKibanaDataSource[dsModel](
			entitycore.ComponentKibana,
			"security_entity_store_entities",
			getDataSourceSchema,
			readEntityStoreEntitiesDataSource,
		),
	}
}

func (d *entitiesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	d.inner.Metadata(ctx, req, resp)
}

func (d *entitiesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	d.inner.Schema(ctx, req, resp)
}

func (d *entitiesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	d.inner.Read(ctx, req, resp)
}

func (d *entitiesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if c, ok := d.inner.(datasource.DataSourceWithConfigure); ok {
		c.Configure(ctx, req, resp)
	}
}

func (d *entitiesDataSource) ValidateConfig(ctx context.Context, req datasource.ValidateConfigRequest, resp *datasource.ValidateConfigResponse) {
	var model dsModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	hasPageMode := !model.SortField.IsNull() || !model.SortOrder.IsNull() || !model.Page.IsNull() || !model.PerPage.IsNull() || !model.FilterQuery.IsNull()

	hasCursorMode := !model.Filter.IsNull() || !model.Size.IsNull() || !model.SearchAfter.IsNull() || !model.Source.IsNull() || !model.Fields.IsNull()

	if hasPageMode && hasCursorMode {
		resp.Diagnostics.AddError(
			"Mixed pagination modes",
			"Page-mode parameters (sort_field, sort_order, page, per_page, filter_query) cannot be combined with cursor-mode parameters (filter, size, search_after, source, fields).",
		)
	}

	if !model.EntityID.IsNull() && !model.EntityID.IsUnknown() {
		if !model.Filter.IsNull() && !model.Filter.IsUnknown() {
			resp.Diagnostics.AddAttributeError(
				path.Root("entity_id"),
				"entity_id conflicts with filter",
				"entity_id cannot be combined with filter.",
			)
		}
		if !model.FilterQuery.IsNull() && !model.FilterQuery.IsUnknown() {
			resp.Diagnostics.AddAttributeError(
				path.Root("entity_id"),
				"entity_id conflicts with filter_query",
				"entity_id cannot be combined with filter_query.",
			)
		}
	}
}
