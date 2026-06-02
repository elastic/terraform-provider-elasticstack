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
)

var (
	_ datasource.DataSourceWithValidateConfig = &entitiesDataSource{}
)

type entitiesDataSource struct {
	datasource.DataSource
}

func NewDataSource() datasource.DataSource {
	return &entitiesDataSource{
		DataSource: entitycore.NewKibanaDataSource[dsModel](
			entitycore.ComponentKibana,
			"security_entity_store_entities",
			getDataSourceSchema,
			readEntityStoreEntitiesDataSource,
		),
	}
}

func (d *entitiesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if c, ok := d.DataSource.(datasource.DataSourceWithConfigure); ok {
		c.Configure(ctx, req, resp)
	}
}

func (d *entitiesDataSource) ValidateConfig(ctx context.Context, req datasource.ValidateConfigRequest, resp *datasource.ValidateConfigResponse) {
	var model dsModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check for mixed pagination modes (cannot use page-mode + cursor-mode together)
	hasPageMode := !model.SortField.IsNull() && !model.SortField.IsUnknown() ||
		!model.SortOrder.IsNull() && !model.SortOrder.IsUnknown() ||
		!model.Page.IsNull() && !model.Page.IsUnknown() ||
		!model.PerPage.IsNull() && !model.PerPage.IsUnknown() ||
		!model.FilterQuery.IsNull() && !model.FilterQuery.IsUnknown()

	hasCursorMode := !model.Filter.IsNull() && !model.Filter.IsUnknown() ||
		!model.Size.IsNull() && !model.Size.IsUnknown() ||
		!model.SearchAfter.IsNull() && !model.SearchAfter.IsUnknown() ||
		!model.Source.IsNull() && !model.Source.IsUnknown() ||
		!model.Fields.IsNull() && !model.Fields.IsUnknown()

	if hasPageMode && hasCursorMode {
		resp.Diagnostics.AddError(
			"Mixed pagination modes",
			"Page-mode parameters (sort_field, sort_order, page, per_page, filter_query) cannot be combined with cursor-mode parameters (filter, size, search_after, source, fields).",
		)
	}
}
