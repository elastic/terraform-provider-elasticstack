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

package spaces

import "github.com/hashicorp/terraform-plugin-framework/types"

// dataSourceModel maps the data source schema data.
type dataSourceModel struct {
	ID     types.String `tfsdk:"id"`
	Spaces []model      `tfsdk:"spaces"`
}

// model maps spaces schema data.
type model struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	DisabledFeatures types.List   `tfsdk:"disabled_features"`
	Initials         types.String `tfsdk:"initials"`
	Color            types.String `tfsdk:"color"`
	ImageURL         types.String `tfsdk:"image_url"`
	Solution         types.String `tfsdk:"solution"`
}
