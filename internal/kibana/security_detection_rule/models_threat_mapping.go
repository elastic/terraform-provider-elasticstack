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

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Helper function to process threat mapping configuration for threat match rules
func (d Data) threatMappingToAPI(ctx context.Context) (kbapi.SecurityDetectionsAPIThreatMapping, diag.Diagnostics) {
	var diags diag.Diagnostics

	threatMapping := make([]TfDataItem, len(d.ThreatMapping.Elements()))

	threatMappingDiags := d.ThreatMapping.ElementsAs(ctx, &threatMapping, false)
	if threatMappingDiags.HasError() {
		diags.Append(threatMappingDiags...)
		return nil, diags
	}

	apiThreatMapping := make(kbapi.SecurityDetectionsAPIThreatMapping, 0)
	for _, mapping := range threatMapping {
		if !typeutils.IsKnown(mapping.Entries) {
			continue
		}

		entries := make([]TfDataItemEntry, len(mapping.Entries.Elements()))
		entryDiag := mapping.Entries.ElementsAs(ctx, &entries, false)
		diags = append(diags, entryDiag...)

		apiThreatMappingEntries := make([]kbapi.SecurityDetectionsAPIThreatMappingEntry, 0)
		for _, entry := range entries {

			apiMapping := kbapi.SecurityDetectionsAPIThreatMappingEntry{
				Field: entry.Field.ValueString(),
				Type:  kbapi.SecurityDetectionsAPIThreatMappingEntryType(entry.Type.ValueString()),
				Value: entry.Value.ValueString(),
			}
			apiThreatMappingEntries = append(apiThreatMappingEntries, apiMapping)

		}

		apiThreatMapping = append(apiThreatMapping, struct {
			Entries []kbapi.SecurityDetectionsAPIThreatMappingEntry `json:"entries"`
		}{Entries: apiThreatMappingEntries})
	}

	return apiThreatMapping, diags
}

// convertThreatMappingToModel converts kbapi.SecurityDetectionsAPIThreatMapping to the terraform model
func convertThreatMappingToModel(ctx context.Context, apiThreatMappings kbapi.SecurityDetectionsAPIThreatMapping) (types.List, diag.Diagnostics) {
	var threatMappings []TfDataItem

	for _, apiMapping := range apiThreatMappings {
		var entries []TfDataItemEntry

		for _, apiEntry := range apiMapping.Entries {
			entries = append(entries, TfDataItemEntry{
				Field: types.StringValue(apiEntry.Field),
				Type:  types.StringValue(string(apiEntry.Type)),
				Value: types.StringValue(apiEntry.Value),
			})
		}

		entriesListValue, diags := types.ListValueFrom(ctx, getThreatMappingEntryElementType(), entries)
		if diags.HasError() {
			return types.ListNull(getThreatMappingElementType()), diags
		}

		threatMappings = append(threatMappings, TfDataItem{
			Entries: entriesListValue,
		})
	}

	listValue, diags := types.ListValueFrom(ctx, getThreatMappingElementType(), threatMappings)
	return listValue, diags
}
