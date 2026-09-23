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

package spacesettings

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var minVersionSpaceSettings = version.Must(version.NewVersion("9.1.0"))

type spaceSettingsModel struct {
	entitycore.ResourceTimeoutsField
	ID                       types.String `tfsdk:"id"`
	KibanaConnection         types.List   `tfsdk:"kibana_connection"`
	SpaceID                  types.String `tfsdk:"space_id"`
	AllowedNamespacePrefixes types.Set    `tfsdk:"allowed_namespace_prefixes"` // > string
	ManagedBy                types.String `tfsdk:"managed_by"`
}

func (m spaceSettingsModel) GetID() types.String             { return m.ID }
func (m spaceSettingsModel) GetResourceID() types.String     { return m.SpaceID }
func (m spaceSettingsModel) GetSpaceID() types.String        { return m.SpaceID }
func (m spaceSettingsModel) GetKibanaConnection() types.List { return m.KibanaConnection }

func (m *spaceSettingsModel) populateFromAPI(ctx context.Context, spaceID string, settings *fleet.SpaceSettings) (diags diag.Diagnostics) {
	m.ID = types.StringValue(spaceID)
	m.SpaceID = types.StringValue(spaceID)
	m.AllowedNamespacePrefixes = typeutils.SetValueFrom(ctx, nonNil(settings.AllowedNamespacePrefixes), types.StringType, path.Root("allowed_namespace_prefixes"), &diags)
	m.ManagedBy = types.StringPointerValue(settings.ManagedBy)
	return diags
}

func (m spaceSettingsModel) GetVersionRequirements(_ context.Context) ([]entitycore.VersionRequirement, diag.Diagnostics) {
	return []entitycore.VersionRequirement{{
		MinVersion:   *minVersionSpaceSettings,
		ErrorMessage: fmt.Sprintf("elasticstack_fleet_space_settings requires server version %s or higher", minVersionSpaceSettings),
	}}, nil
}

func (m spaceSettingsModel) prefixesToWrite(ctx context.Context) ([]string, diag.Diagnostics) {
	var diags diag.Diagnostics
	prefixes := typeutils.SetTypeAs[string](ctx, m.AllowedNamespacePrefixes, path.Root("allowed_namespace_prefixes"), &diags)
	return nonNil(prefixes), diags
}

// An empty set must be written as [] and stored as an empty set, never null.
func nonNil(values []string) []string {
	if values == nil {
		return []string{}
	}
	return values
}
