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

package slm

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	esclients "github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	fwtypes "github.com/hashicorp/terraform-plugin-framework/types"
)

func writeSlm(ctx context.Context, client *esclients.ElasticsearchScopedClient, resourceID string, data Data) (Data, diag.Diagnostics) {
	var diags diag.Diagnostics

	id, sdkDiags := client.ID(ctx, resourceID)
	diags.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if diags.HasError() {
		var zero Data
		return zero, diags
	}

	var slmPolicy elasticsearch.SlmPolicy
	slmPolicy.Name = data.SnapshotName.ValueString()
	slmPolicy.Repository = data.Repository.ValueString()
	slmPolicy.Schedule = data.Schedule.ValueString()

	// Build retention
	var retention elasticsearch.SlmRetention
	hasRetention := false
	if !data.ExpireAfter.IsNull() && !data.ExpireAfter.IsUnknown() && data.ExpireAfter.ValueString() != "" {
		v := data.ExpireAfter.ValueString()
		retention.ExpireAfter = &v
		hasRetention = true
	}
	if !data.MaxCount.IsNull() && !data.MaxCount.IsUnknown() {
		v := int(data.MaxCount.ValueInt64())
		retention.MaxCount = &v
		hasRetention = true
	}
	if !data.MinCount.IsNull() && !data.MinCount.IsUnknown() {
		v := int(data.MinCount.ValueInt64())
		retention.MinCount = &v
		hasRetention = true
	}
	if hasRetention {
		slmPolicy.Retention = &retention
	}

	// Build config
	var cfg elasticsearch.SlmConfig
	cfg.ExpandWildcards = data.ExpandWildcards.ValueString()
	ignoreUnavailable := data.IgnoreUnavailable.ValueBool()
	cfg.IgnoreUnavailable = &ignoreUnavailable
	includeGlobalState := data.IncludeGlobalState.ValueBool()
	cfg.IncludeGlobalState = &includeGlobalState
	partial := data.Partial.ValueBool()
	cfg.Partial = &partial

	// Indices
	if !data.Indices.IsNull() && !data.Indices.IsUnknown() {
		var indices []string
		diags.Append(data.Indices.ElementsAs(ctx, &indices, false)...)
		if diags.HasError() {
			var zero Data
			return zero, diags
		}
		cfg.Indices = indices
	}

	// FeatureStates
	if !data.FeatureStates.IsNull() && !data.FeatureStates.IsUnknown() {
		var featureStates []string
		diags.Append(data.FeatureStates.ElementsAs(ctx, &featureStates, false)...)
		if diags.HasError() {
			var zero Data
			return zero, diags
		}
		cfg.FeatureStates = featureStates
	}

	// Metadata
	if !data.Metadata.IsNull() && !data.Metadata.IsUnknown() {
		metaStr := data.Metadata.ValueString()
		if metaStr != "" {
			var metadata map[string]any
			if err := json.NewDecoder(strings.NewReader(metaStr)).Decode(&metadata); err != nil {
				diags.AddError("Failed to decode metadata", err.Error())
				var zero Data
				return zero, diags
			}
			metaRaw := make(types.Metadata)
			for k, val := range metadata {
				raw, err := json.Marshal(val)
				if err != nil {
					diags.AddError("Failed to encode metadata value", err.Error())
					var zero Data
					return zero, diags
				}
				metaRaw[k] = raw
			}
			cfg.Metadata = metaRaw
		}
	}

	slmPolicy.Config = &cfg

	diags.Append(diagutil.FrameworkDiagsFromSDK(elasticsearch.PutSlm(ctx, client, resourceID, &slmPolicy))...)
	if diags.HasError() {
		var zero Data
		return zero, diags
	}

	data.ID = fwtypes.StringValue(id.String())
	return data, diags
}
