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

package securitydetectionrule

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// filtersToAPI converts the Terraform filters field to the API type
func (d Data) filtersToAPI(ctx context.Context) (*kbapi.SecurityDetectionsAPIRuleFilterArray, diag.Diagnostics) {
	var diags diag.Diagnostics
	_ = ctx

	if !typeutils.IsKnown(d.Filters) {
		return nil, diags
	}

	// Unmarshal the JSON string to []interface{}
	var filters kbapi.SecurityDetectionsAPIRuleFilterArray
	unmarshalDiags := d.Filters.Unmarshal(&filters)
	diags.Append(unmarshalDiags...)

	if diags.HasError() {
		return nil, diags
	}

	return &filters, diags
}

// threatFiltersToAPI converts the Terraform threat_filters list (JSON strings) into the API type ([]interface{}).
func (d Data) threatFiltersToAPI(ctx context.Context) (*kbapi.SecurityDetectionsAPIThreatFilters, diag.Diagnostics) {
	var diags diag.Diagnostics

	if !typeutils.IsKnown(d.ThreatFilters) {
		return nil, diags
	}

	filtersJSON := typeutils.ListTypeAs[string](ctx, d.ThreatFilters, path.Root("threat_filters"), &diags)
	if diags.HasError() {
		return nil, diags
	}

	apiThreatFilters := make(kbapi.SecurityDetectionsAPIThreatFilters, 0, len(filtersJSON))
	for i, filterStr := range filtersJSON {
		var filter any
		if err := json.Unmarshal([]byte(filterStr), &filter); err != nil {
			diags.AddError(
				"Invalid threat_filters JSON",
				fmt.Sprintf("threat_filters[%d] is not valid JSON: %s", i, err.Error()),
			)
			continue
		}
		apiThreatFilters = append(apiThreatFilters, filter)
	}

	return &apiThreatFilters, diags
}

func (d *Data) updateFiltersFromAPI(ctx context.Context, apiFilters *kbapi.SecurityDetectionsAPIRuleFilterArray) diag.Diagnostics {
	var diags diag.Diagnostics
	_ = ctx

	if apiFilters == nil || len(*apiFilters) == 0 {
		d.Filters = jsontypes.NewNormalizedNull()
		return diags
	}

	// Marshal the []interface{} to JSON string
	jsonBytes, err := json.Marshal(*apiFilters)
	if err != nil {
		diags.AddError("Failed to marshal filters", err.Error())
		return diags
	}

	// Create a NormalizedValue from the JSON string
	d.Filters = jsontypes.NewNormalizedValue(string(jsonBytes))
	return diags
}

func (d *Data) updateThreatFiltersFromAPI(ctx context.Context, apiThreatFilters *kbapi.SecurityDetectionsAPIThreatFilters) diag.Diagnostics {
	var diags diag.Diagnostics
	_ = ctx

	if apiThreatFilters == nil {
		d.ThreatFilters = types.ListNull(types.StringType)
		return diags
	}

	if len(*apiThreatFilters) == 0 {
		d.ThreatFilters = typeutils.StringsToListMust(nil)
		return diags
	}

	filters := make([]string, 0, len(*apiThreatFilters))
	for i, filter := range *apiThreatFilters {
		jsonBytes, err := json.Marshal(filter)
		if err != nil {
			diags.AddError("Failed to marshal threat_filters item", fmt.Sprintf("threat_filters[%d]: %s", i, err.Error()))
			continue
		}
		filters = append(filters, string(jsonBytes))
	}

	d.ThreatFilters = typeutils.ListValueFrom(ctx, filters, types.StringType, path.Root("threat_filters"), &diags)
	return diags
}
