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

import (
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// dataSourceModel maps the data source schema data.
type dataSourceModel struct {
	entitycore.KibanaConnectionField
	ID     types.String `tfsdk:"id"`
	Spaces []SpaceModel `tfsdk:"spaces"`
}

func (m dataSourceModel) GetID() types.String         { return m.ID }
func (m dataSourceModel) GetResourceID() types.String { return types.StringValue("spaces") }
func (m dataSourceModel) GetSpaceID() types.String    { return types.StringValue("default") }

// SpaceModel maps spaces schema data for the list data source.
type SpaceModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	DisabledFeatures types.List   `tfsdk:"disabled_features"`
	Initials         types.String `tfsdk:"initials"`
	Color            types.String `tfsdk:"color"`
	ImageURL         types.String `tfsdk:"image_url"`
	Solution         types.String `tfsdk:"solution"`
}

type resourceModel struct {
	entitycore.KibanaConnectionField
	ID               types.String `tfsdk:"id"`
	SpaceID          types.String `tfsdk:"space_id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	DisabledFeatures types.Set    `tfsdk:"disabled_features"`
	Initials         types.String `tfsdk:"initials"`
	Color            types.String `tfsdk:"color"`
	ImageURL         types.String `tfsdk:"image_url"`
	Solution         types.String `tfsdk:"solution"`
}

func (m resourceModel) GetID() types.String         { return m.ID }
func (m resourceModel) GetResourceID() types.String { return m.ID }
func (m resourceModel) GetSpaceID() types.String    { return types.StringValue("default") }

var spaceSolutionMinVersion = version.Must(version.NewVersion("8.16.0"))

const spaceSolutionVersionErrorMessage = "solution field is not supported in this version of the Elastic Stack. Solution field requires 8.16.0 or higher"

func (m resourceModel) GetVersionRequirements() ([]entitycore.VersionRequirement, diag.Diagnostics) {
	if m.Solution.IsNull() || m.Solution.IsUnknown() || m.Solution.ValueString() == "" {
		return nil, nil
	}
	return []entitycore.VersionRequirement{
		{
			MinVersion:   *spaceSolutionMinVersion,
			ErrorMessage: spaceSolutionVersionErrorMessage,
		},
	}, nil
}
