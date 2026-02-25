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
	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type prebuiltRuleModel struct {
	ID                    types.String `tfsdk:"id"`
	SpaceID               types.String `tfsdk:"space_id"`
	RulesInstalled        types.Int64  `tfsdk:"rules_installed"`
	RulesNotInstalled     types.Int64  `tfsdk:"rules_not_installed"`
	RulesNotUpdated       types.Int64  `tfsdk:"rules_not_updated"`
	TimelinesInstalled    types.Int64  `tfsdk:"timelines_installed"`
	TimelinesNotInstalled types.Int64  `tfsdk:"timelines_not_installed"`
	TimelinesNotUpdated   types.Int64  `tfsdk:"timelines_not_updated"`
}

func (model *prebuiltRuleModel) populateFromStatus(status *kbapi.ReadPrebuiltRulesAndTimelinesStatusResponse) {
	model.RulesInstalled = types.Int64Value(int64(status.JSON200.RulesInstalled))
	model.RulesNotInstalled = types.Int64Value(int64(status.JSON200.RulesNotInstalled))
	model.RulesNotUpdated = types.Int64Value(int64(status.JSON200.RulesNotUpdated))
	model.TimelinesInstalled = types.Int64Value(int64(status.JSON200.TimelinesInstalled))
	model.TimelinesNotInstalled = types.Int64Value(int64(status.JSON200.TimelinesNotInstalled))
	model.TimelinesNotUpdated = types.Int64Value(int64(status.JSON200.TimelinesNotUpdated))
}
