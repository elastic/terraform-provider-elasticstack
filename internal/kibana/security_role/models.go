// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
//
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

package security_role

import (
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	minSupportedRemoteIndicesVersion = version.Must(version.NewVersion("8.10.0"))
	minSupportedDescriptionVersion   = version.Must(version.NewVersion("8.15.0"))
)

type resourceModel struct {
	entitycore.KibanaConnectionField
	ID            types.String         `tfsdk:"id"`
	Name          types.String         `tfsdk:"name"`
	Description   types.String         `tfsdk:"description"`
	Metadata      jsontypes.Normalized `tfsdk:"metadata"`
	Elasticsearch types.Set            `tfsdk:"elasticsearch"`
	Kibana        types.Set            `tfsdk:"kibana"`
}

func (m resourceModel) GetID() types.String { return m.ID }

func (m resourceModel) GetResourceID() types.String {
	if m.Name.IsNull() || m.Name.IsUnknown() || m.Name.ValueString() == "" {
		return m.ID
	}
	return m.Name
}

func (m resourceModel) GetSpaceID() types.String {
	return types.StringValue("")
}

// IsUnscopedSpace reports that Kibana role APIs are not space-scoped.
func (resourceModel) IsUnscopedSpace() bool { return true }

var _ entitycore.KibanaUnscopedSpace = resourceModel{}
var _ entitycore.WithVersionRequirements = resourceModel{}

// GetVersionRequirements satisfies [entitycore.WithVersionRequirements].
func (m resourceModel) GetVersionRequirements() ([]entitycore.VersionRequirement, diag.Diagnostics) {
	var reqs []entitycore.VersionRequirement
	var diags diag.Diagnostics

	if m.requiresDescriptionVersion() {
		reqs = append(reqs, entitycore.VersionRequirement{
			MinVersion:   *minSupportedDescriptionVersion,
			ErrorMessage: fmt.Sprintf(`'description' is supported only for Kibana v%s and above`, minSupportedDescriptionVersion.String()),
		})
	}

	if m.requiresRemoteIndicesVersion() {
		reqs = append(reqs, entitycore.VersionRequirement{
			MinVersion:   *minSupportedRemoteIndicesVersion,
			ErrorMessage: fmt.Sprintf(`'remote_indices' is supported only for Kibana v%s and above`, minSupportedRemoteIndicesVersion.String()),
		})
	}

	return reqs, diags
}

func (m resourceModel) requiresDescriptionVersion() bool {
	return typeutils.IsKnown(m.Description) && m.Description.ValueString() != ""
}

func (m resourceModel) requiresRemoteIndicesVersion() bool {
	if m.Elasticsearch.IsNull() || m.Elasticsearch.IsUnknown() {
		return false
	}
	for _, elem := range m.Elasticsearch.Elements() {
		obj, ok := elem.(types.Object)
		if !ok {
			continue
		}
		attrs := obj.Attributes()
		ri, ok := attrs["remote_indices"].(types.Set)
		if !ok || ri.IsNull() || ri.IsUnknown() {
			continue
		}
		if len(ri.Elements()) > 0 {
			return true
		}
	}
	return false
}

type dataSourceModel struct {
	entitycore.KibanaConnectionField
	Name          types.String         `tfsdk:"name"`
	Description   types.String         `tfsdk:"description"`
	Metadata      jsontypes.Normalized `tfsdk:"metadata"`
	Elasticsearch types.Set            `tfsdk:"elasticsearch"`
	Kibana        types.Set            `tfsdk:"kibana"`
}
