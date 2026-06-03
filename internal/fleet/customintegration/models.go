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

package customintegration

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	goversion "github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var minVersionCustomPackageGet = goversion.Must(goversion.NewVersion("8.2.0"))

type customIntegrationModel struct {
	ID                        types.String   `tfsdk:"id"`
	KibanaConnection          types.List     `tfsdk:"kibana_connection"`
	PackagePath               types.String   `tfsdk:"package_path"`
	PackageName               types.String   `tfsdk:"package_name"`
	PackageVersion            types.String   `tfsdk:"package_version"`
	Checksum                  types.String   `tfsdk:"checksum"`
	IgnoreMappingUpdateErrors types.Bool     `tfsdk:"ignore_mapping_update_errors"`
	SkipDataStreamRollover    types.Bool     `tfsdk:"skip_data_stream_rollover"`
	SkipDestroy               types.Bool     `tfsdk:"skip_destroy"`
	SpaceID                   types.String   `tfsdk:"space_id"`
	Timeouts                  timeouts.Value `tfsdk:"timeouts"`
}

func (m customIntegrationModel) GetID() types.String {
	// Return a composite ID so the envelope's resolveKibanaResourceIdentity
	// does not misinterpret the raw package ID (name/version) as a space/resource
	// pair. The format is "<space>/<package_id>".
	rawID := m.ID.ValueString()
	if rawID == "" {
		return m.ID
	}
	spaceID := m.GetSpaceID().ValueString()
	return types.StringValue(fmt.Sprintf("%s/%s", spaceID, rawID))
}

func (m customIntegrationModel) GetResourceID() types.String {
	if !m.PackageName.IsNull() && !m.PackageName.IsUnknown() &&
		!m.PackageVersion.IsNull() && !m.PackageVersion.IsUnknown() {
		return types.StringValue(getPackageID(m.PackageName.ValueString(), m.PackageVersion.ValueString()))
	}
	return m.ID
}

func (m customIntegrationModel) GetSpaceID() types.String {
	// Return "default" when space_id is unset so the envelope's
	// validateSpaceID accepts the plan. Fleet APIs treat "" and "default"
	// identically (BuildSpaceAwarePath), so callbacks remain backward-
	// compatible when they use model.SpaceID.ValueString().
	if m.SpaceID.IsNull() || m.SpaceID.IsUnknown() || m.SpaceID.ValueString() == "" {
		return types.StringValue("default")
	}
	return m.SpaceID
}

func (m customIntegrationModel) GetKibanaConnection() types.List {
	return m.KibanaConnection
}

var customIntegrationVersionReqs = []entitycore.VersionRequirement{{
	MinVersion:   *minVersionCustomPackageGet,
	ErrorMessage: "elasticstack_fleet_custom_integration requires Kibana 8.2.0 or later.",
}}

func (m customIntegrationModel) GetVersionRequirements(_ context.Context) ([]entitycore.VersionRequirement, diag.Diagnostics) {
	return customIntegrationVersionReqs, nil
}

func getPackageID(name string, version string) string {
	return fmt.Sprintf("%s/%s", name, version)
}
