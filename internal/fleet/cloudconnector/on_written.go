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

package cloudconnector

import (
	"context"
	"sort"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func onWrittenCloudConnector(
	ctx context.Context,
	_ *clients.KibanaScopedClient,
	_ cloudConnectorModel,
	config cloudConnectorModel,
	privateState any,
) diag.Diagnostics {
	var diags diag.Diagnostics

	priv, ok := privateState.(privateData)
	if !ok || priv == nil {
		return diags
	}

	hasher := cloudConnectorHasher()
	entries, entryDiags := writeOnlyEntriesFromConfig(ctx, config)
	diags.Append(entryDiags...)
	if diags.HasError() {
		return diags
	}

	hasAWSExternalID := false
	currentVarKeys := make([]string, 0)
	for _, entry := range entries {
		if entry.attributePath == writeOnlyAttributeAWSExternalID {
			hasAWSExternalID = true
		}
		hash, err := hasher.Compute(entry.value.ValueString())
		if err != nil {
			diags.AddError(
				"Failed to hash write-only attribute",
				err.Error(),
			)
			return diags
		}
		diags.Append(priv.SetKey(ctx, hasher.PrivateStateKey(entry.attributePath), hash)...)
		if diags.HasError() {
			return diags
		}
		if entry.attributePath != writeOnlyAttributeAWSExternalID {
			varKey := entry.attributePath[len("vars.") : len(entry.attributePath)-len(".secret_value")]
			currentVarKeys = append(currentVarKeys, varKey)
		}
	}

	if !hasAWSExternalID {
		diags.Append(priv.SetKey(ctx, awsExternalIDPrivateStateKey(), nil)...)
		if diags.HasError() {
			return diags
		}
	}

	sort.Strings(currentVarKeys)

	indexBytes, indexDiags := priv.GetKey(ctx, varsSecretIndexPrivateStateKey)
	diags.Append(indexDiags...)
	if diags.HasError() {
		return diags
	}
	previousVarKeys, parseDiags := parseVarsSecretIndex(indexBytes)
	diags.Append(parseDiags...)
	if diags.HasError() {
		return diags
	}

	currentVarKeySet := make(map[string]struct{}, len(currentVarKeys))
	for _, key := range currentVarKeys {
		currentVarKeySet[key] = struct{}{}
	}
	for _, key := range previousVarKeys {
		if _, stillPresent := currentVarKeySet[key]; stillPresent {
			continue
		}
		diags.Append(priv.SetKey(ctx, varsSecretValuePrivateStateKey(key), nil)...)
		if diags.HasError() {
			return diags
		}
	}

	if len(currentVarKeys) == 0 {
		diags.Append(priv.SetKey(ctx, varsSecretIndexPrivateStateKey, nil)...)
		return diags
	}

	indexData, marshalDiags := marshalVarsSecretIndex(currentVarKeys)
	diags.Append(marshalDiags...)
	if diags.HasError() {
		return diags
	}
	diags.Append(priv.SetKey(ctx, varsSecretIndexPrivateStateKey, indexData)...)

	return diags
}
