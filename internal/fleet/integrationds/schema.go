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

package integrationds

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func (d *integrationDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema.Description = `This data source provides information about a Fleet integration package. Currently,
the data source will retrieve the latest available version of the package. Version
selection is determined by the Fleet API, which is currently based on semantic
versioning.

By default, the highest GA release version will be selected. If a
package is not GA (the version is below 1.0.0) or if a new non-GA version of the
package is to be selected (i.e., the GA version of the package is 1.5.0, but there's
a new 1.5.1-beta version available), then the ` + "`prerelease`" + ` parameter in the plan
should be set to ` + "`true`."
	resp.Schema.Attributes = map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "The ID of this resource.",
			Computed:    true,
		},
		"name": schema.StringAttribute{
			Description: "The integration package name.",
			Required:    true,
		},
		"prerelease": schema.BoolAttribute{
			Description: "Include prerelease packages.",
			Optional:    true,
		},
		"version": schema.StringAttribute{
			Description: "The integration package version.",
			Computed:    true,
		},
	}
}
