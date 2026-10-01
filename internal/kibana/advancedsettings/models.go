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

package advancedsettings

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var minGlobalSettingsVersion = version.Must(version.NewVersion("8.7.0"))

type advancedSettingsModel struct {
	entitycore.ResourceTimeoutsField
	ID               types.String `tfsdk:"id"`
	KibanaConnection types.List   `tfsdk:"kibana_connection"`
	SpaceID          types.String `tfsdk:"space_id"`
	Global           types.Bool   `tfsdk:"global"`
	Settings         types.Map    `tfsdk:"settings"`
}

// scopeID returns the identifier of the settings scope managed by the model.
func (m advancedSettingsModel) scopeID() string {
	if m.Global.ValueBool() {
		return globalScopeID
	}
	return m.SpaceID.ValueString()
}

func (m advancedSettingsModel) GetID() types.String             { return m.ID }
func (m advancedSettingsModel) GetResourceID() types.String     { return types.StringValue(m.scopeID()) }
func (m advancedSettingsModel) GetSpaceID() types.String        { return m.SpaceID }
func (m advancedSettingsModel) GetKibanaConnection() types.List { return m.KibanaConnection }

// IsUnscopedSpace reports whether the model manages the global scope, which
// is not tied to a space.
func (m advancedSettingsModel) IsUnscopedSpace() bool { return m.Global.ValueBool() }

func (m advancedSettingsModel) GetVersionRequirements(_ context.Context) ([]entitycore.VersionRequirement, diag.Diagnostics) {
	if !m.Global.ValueBool() {
		return nil, nil
	}
	return []entitycore.VersionRequirement{
		entitycore.NewAttributeVersionRequirement(
			path.Root(attrGlobal),
			*minGlobalSettingsVersion,
			"Global advanced settings require Kibana 8.7.0 or later.",
		),
	}, nil
}

var (
	_ entitycore.KibanaResourceModel     = advancedSettingsModel{}
	_ entitycore.KibanaUnscopedSpace     = advancedSettingsModel{}
	_ entitycore.WithVersionRequirements = advancedSettingsModel{}
)
