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

package prebuiltrules

import (
	"context"
	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type prebuiltRuleModel struct {
	entitycore.ResourceTimeoutsField
	ID                    types.String `tfsdk:"id"`
	SpaceID               types.String `tfsdk:"space_id"`
	KibanaConnection      types.List   `tfsdk:"kibana_connection"`
	RulesInstalled        types.Int64  `tfsdk:"rules_installed"`
	RulesNotInstalled     types.Int64  `tfsdk:"rules_not_installed"`
	RulesNotUpdated       types.Int64  `tfsdk:"rules_not_updated"`
	TimelinesInstalled    types.Int64  `tfsdk:"timelines_installed"`
	TimelinesNotInstalled types.Int64  `tfsdk:"timelines_not_installed"`
	TimelinesNotUpdated   types.Int64  `tfsdk:"timelines_not_updated"`
}

func (m prebuiltRuleModel) GetID() types.String             { return m.ID }
func (m prebuiltRuleModel) GetResourceID() types.String     { return m.SpaceID }
func (m prebuiltRuleModel) GetSpaceID() types.String        { return m.SpaceID }
func (m prebuiltRuleModel) GetKibanaConnection() types.List { return m.KibanaConnection }

var _ entitycore.KibanaResourceModel = prebuiltRuleModel{}

var minSupportedVersion = version.Must(version.NewVersion("8.0.0"))

func (m prebuiltRuleModel) GetVersionRequirements(_ context.Context) ([]entitycore.VersionRequirement, diag.Diagnostics) {
	return []entitycore.VersionRequirement{
		{
			MinVersion:   *minSupportedVersion,
			ErrorMessage: "Prebuilt rules are not supported until Elastic Stack v8.0.0. Upgrade the target server to use this resource",
		},
	}, nil
}

func (m *prebuiltRuleModel) populateFromStatus(status *kbapi.ReadPrebuiltRulesAndTimelinesStatusResponse) {
	m.RulesInstalled = types.Int64Value(int64(status.JSON200.RulesInstalled))
	m.RulesNotInstalled = types.Int64Value(int64(status.JSON200.RulesNotInstalled))
	m.RulesNotUpdated = types.Int64Value(int64(status.JSON200.RulesNotUpdated))
	m.TimelinesInstalled = types.Int64Value(int64(status.JSON200.TimelinesInstalled))
	m.TimelinesNotInstalled = types.Int64Value(int64(status.JSON200.TimelinesNotInstalled))
	m.TimelinesNotUpdated = types.Int64Value(int64(status.JSON200.TimelinesNotUpdated))
}
