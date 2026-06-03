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

package securityenablerule

import (
	"context"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type enableRuleModel struct {
	ID               types.String `tfsdk:"id"`
	KibanaConnection types.List   `tfsdk:"kibana_connection"`
	SpaceID          types.String `tfsdk:"space_id"`
	Key              types.String `tfsdk:"key"`
	Value            types.String `tfsdk:"value"`
	DisableOnDestroy types.Bool   `tfsdk:"disable_on_destroy"`
	AllRulesEnabled  types.Bool   `tfsdk:"all_rules_enabled"`
}

func (m enableRuleModel) GetID() types.String             { return m.ID }
func (m enableRuleModel) GetResourceID() types.String     { return m.Key }
func (m enableRuleModel) GetSpaceID() types.String        { return m.SpaceID }
func (m enableRuleModel) GetKibanaConnection() types.List { return m.KibanaConnection }

var _ entitycore.KibanaResourceModel = enableRuleModel{}

var minSupportedVersion = version.Must(version.NewVersion("8.11.0"))

func (m enableRuleModel) GetVersionRequirements(_ context.Context) ([]entitycore.VersionRequirement, diag.Diagnostics) {
	return []entitycore.VersionRequirement{
		{
			MinVersion:   *minSupportedVersion,
			ErrorMessage: "Security detection rules bulk actions are not supported until Elastic Stack v8.11.0. Upgrade the target server to use this resource",
		},
	}, nil
}
