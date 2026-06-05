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

package integration

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Terraform schema attribute keys reused across the schema definition, the
// V0 state-upgrade schema, and debug logging.
const (
	attrName    = "name"
	attrVersion = "version"
)

type integrationModel struct {
	ID                        types.String `tfsdk:"id"`
	KibanaConnection          types.List   `tfsdk:"kibana_connection"`
	Name                      types.String `tfsdk:"name"`
	Version                   types.String `tfsdk:"version"`
	Force                     types.Bool   `tfsdk:"force"`
	Prerelease                types.Bool   `tfsdk:"prerelease"`
	IgnoreMappingUpdateErrors types.Bool   `tfsdk:"ignore_mapping_update_errors"`
	SkipDataStreamRollover    types.Bool   `tfsdk:"skip_data_stream_rollover"`
	IgnoreConstraints         types.Bool   `tfsdk:"ignore_constraints"`
	SkipDestroy               types.Bool   `tfsdk:"skip_destroy"`
	SpaceID                   types.String `tfsdk:"space_id"`
}

func (m integrationModel) GetID() types.String {
	return m.ID
}

func (m integrationModel) GetResourceID() types.String {
	return types.StringValue(getPackageID(m.Name.ValueString(), m.Version.ValueString()))
}

func (m integrationModel) GetSpaceID() types.String {
	if m.SpaceID.IsNull() || m.SpaceID.IsUnknown() {
		return types.StringValue("")
	}
	return m.SpaceID
}

func (m integrationModel) GetKibanaConnection() types.List {
	return m.KibanaConnection
}

func (m integrationModel) IsUnscopedSpace() bool {
	return true
}

var _ entitycore.WithVersionRequirements = integrationModel{}

// GetVersionRequirements satisfies [entitycore.WithVersionRequirements].
func (m integrationModel) GetVersionRequirements(_ context.Context) ([]entitycore.VersionRequirement, diag.Diagnostics) {
	var reqs []entitycore.VersionRequirement

	if typeutils.IsKnown(m.IgnoreMappingUpdateErrors) && m.IgnoreMappingUpdateErrors.ValueBool() {
		reqs = append(reqs, entitycore.VersionRequirement{
			MinVersion:   *MinVersionIgnoreMappingUpdateErrors,
			ErrorMessage: "The 'ignore_mapping_update_errors' parameter requires server version " + MinVersionIgnoreMappingUpdateErrors.String() + " or higher.",
		})
	}

	if typeutils.IsKnown(m.SkipDataStreamRollover) && m.SkipDataStreamRollover.ValueBool() {
		reqs = append(reqs, entitycore.VersionRequirement{
			MinVersion:   *MinVersionSkipDataStreamRollover,
			ErrorMessage: "The 'skip_data_stream_rollover' parameter requires server version " + MinVersionSkipDataStreamRollover.String() + " or higher.",
		})
	}

	return reqs, nil
}

func getPackageID(name string, version string) string {
	hash, _ := typeutils.StringToHash(name + version)
	return *hash
}
