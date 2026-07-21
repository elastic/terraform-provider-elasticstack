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

package managedintegration

import (
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// managedIntegrationFromPackagePolicyReadResponse projects a simplified
// (mapped-format) package_policies GET/PUT body onto
// KibanaHTTPAPIsManagedIntegration for state population. Delete with task 8
// when read/update call ReadManagedIntegration / UpdateManagedIntegration
// directly.
func managedIntegrationFromPackagePolicyReadResponse(data *kbapi.PackagePolicy) (kbapi.KibanaHTTPAPIsManagedIntegration, diag.Diagnostics) {
	var diags diag.Diagnostics
	if data == nil {
		return kbapi.KibanaHTTPAPIsManagedIntegration{}, diags
	}

	mappedInputs, err := data.Inputs.AsPackagePolicyMappedInputs()
	if err != nil {
		mappedInputs = kbapi.PackagePolicyMappedInputs{}
	}

	wire := map[string]any{
		"id":          data.Id,
		"name":        data.Name,
		"description": data.Description,
		"namespace":   data.Namespace,
		"created_at":  data.CreatedAt,
		"created_by":  data.CreatedBy,
		"updated_at":  data.UpdatedAt,
		"updated_by":  data.UpdatedBy,
		"vars":        data.Vars,
		"inputs":      mappedInputs,
	}
	if data.VarGroupSelections != nil {
		wire["var_group_selections"] = data.VarGroupSelections
	}
	if data.AdditionalDatastreamsPermissions != nil {
		wire["additional_datastreams_permissions"] = data.AdditionalDatastreamsPermissions
	}
	if data.GlobalDataTags != nil {
		wire["global_data_tags"] = data.GlobalDataTags
	}
	if data.Package != nil {
		wire["package"] = data.Package
	}

	b, err := json.Marshal(wire)
	if err != nil {
		diags.AddError("Failed to map package policy read response", err.Error())
		return kbapi.KibanaHTTPAPIsManagedIntegration{}, diags
	}

	var out kbapi.KibanaHTTPAPIsManagedIntegration
	if err := json.Unmarshal(b, &out); err != nil {
		diags.AddError("Failed to map package policy read response", err.Error())
		return kbapi.KibanaHTTPAPIsManagedIntegration{}, diags
	}

	return out, diags
}
