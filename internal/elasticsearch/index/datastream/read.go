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

package datastream

import (
	"context"
	"encoding/json"
	"fmt"

	estypes "github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func readDataStream(ctx context.Context, client *clients.ElasticsearchScopedClient, resourceID string, state Data) (Data, bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	ds, sdkDiags := elasticsearch.GetDataStream(ctx, client, resourceID)
	diags.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if diags.HasError() {
		return state, false, diags
	}

	if ds == nil {
		tflog.Warn(ctx, fmt.Sprintf(`Data stream "%s" not found, removing from state`, resourceID))
		return state, false, diags
	}

	state.Name = types.StringValue(ds.Name)
	state.TimestampField = types.StringValue(ds.TimestampField.Name)
	state.Generation = types.Int64Value(int64(ds.Generation))
	state.Status = types.StringValue(ds.Status.String())
	state.Template = types.StringValue(ds.Template)
	state.Hidden = types.BoolValue(ds.Hidden)

	ilmPolicy := ""
	if ds.IlmPolicy != nil {
		ilmPolicy = *ds.IlmPolicy
	}
	state.ILMPolicy = types.StringValue(ilmPolicy)

	system := false
	if ds.System != nil {
		system = *ds.System
	}
	state.System = types.BoolValue(system)

	replicated := false
	if ds.Replicated != nil {
		replicated = *ds.Replicated
	}
	state.Replicated = types.BoolValue(replicated)

	if ds.Meta_ != nil {
		metadataBytes, err := json.Marshal(ds.Meta_)
		if err != nil {
			diags.AddError("Failed to marshal data stream metadata", err.Error())
			return state, false, diags
		}
		state.Metadata = types.StringValue(string(metadataBytes))
	} else {
		state.Metadata = types.StringNull()
	}

	indicesVal, indicesDiags := buildIndicesList(ctx, ds.Indices)
	diags.Append(indicesDiags...)
	if diags.HasError() {
		return state, false, diags
	}
	state.Indices = indicesVal

	return state, true, diags
}

func buildIndicesList(ctx context.Context, esIndices []estypes.DataStreamIndex) (types.List, diag.Diagnostics) {
	models := make([]indexModel, len(esIndices))
	for i, idx := range esIndices {
		models[i] = indexModel{
			IndexName: types.StringValue(idx.IndexName),
			IndexUUID: types.StringValue(idx.IndexUuid),
		}
	}
	return types.ListValueFrom(ctx, indicesElementType(), models)
}
