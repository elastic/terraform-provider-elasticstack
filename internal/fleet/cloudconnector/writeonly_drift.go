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
	"encoding/json"
	"fmt"
	"sync"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/writeonlyhash"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

const (
	cloudConnectorResourceType = "elasticstack_fleet_cloud_connector"

	writeOnlyAttributeAWSExternalID = "aws.external_id"

	// varsSecretIndexPrivateStateKey tracks var map keys that have secret_value
	// hashes in private state. The Plugin Framework private-state API does not
	// support enumerating keys, so this JSON-encoded index is maintained on each
	// write and consulted during ModifyPlan for stale-key drift and cleanup.
	varsSecretIndexPrivateStateKey  = "secret_hash:vars._index"
	varsSecretValueAttributePattern = "vars.%s.secret_value"
)

var (
	cloudConnectorHasherOnce sync.Once
	cloudConnectorHasherInst *writeonlyhash.Hasher
)

func cloudConnectorHasher() *writeonlyhash.Hasher {
	cloudConnectorHasherOnce.Do(func() {
		cloudConnectorHasherInst = writeonlyhash.New(cloudConnectorResourceType)
	})
	return cloudConnectorHasherInst
}

func awsExternalIDPrivateStateKey() string {
	return cloudConnectorHasher().PrivateStateKey(writeOnlyAttributeAWSExternalID)
}

func varsSecretValuePrivateStateKey(varKey string) string {
	return cloudConnectorHasher().PrivateStateKey(fmt.Sprintf(varsSecretValueAttributePattern, varKey))
}

func varsSecretValueAttributePath(varKey string) string {
	return fmt.Sprintf(varsSecretValueAttributePattern, varKey)
}

// privateData mirrors the GetKey/SetKey surface of *privatestate.ProviderData
// so write-only hash logic can run without importing the framework internal package.
type privateData interface {
	GetKey(ctx context.Context, key string) ([]byte, diag.Diagnostics)
	SetKey(ctx context.Context, key string, value []byte) diag.Diagnostics
}

type driftResult struct {
	Changed          bool
	AttributePath    string
	IsImportBaseline bool
}

// detectWriteOnlyDrift compares a configured write-only value against a stored
// bcrypt hash. configValue null/unknown with a stored hash indicates removal
// from config; a set config value with no stored hash is the import-baseline case.
func detectWriteOnlyDrift(hasher *writeonlyhash.Hasher, attributePath string, configValue types.String, storedHash []byte) driftResult {
	if !typeutils.IsKnown(configValue) {
		if len(storedHash) > 0 {
			return driftResult{
				Changed:       true,
				AttributePath: attributePath,
			}
		}
		return driftResult{}
	}

	if len(storedHash) == 0 {
		return driftResult{
			Changed:          true,
			AttributePath:    attributePath,
			IsImportBaseline: true,
		}
	}

	if !hasher.Matches(configValue.ValueString(), storedHash) {
		return driftResult{
			Changed:       true,
			AttributePath: attributePath,
		}
	}

	return driftResult{}
}

type writeOnlyConfigEntry struct {
	attributePath string
	value         types.String
}

func writeOnlyEntriesFromConfig(ctx context.Context, config cloudConnectorModel) ([]writeOnlyConfigEntry, diag.Diagnostics) {
	var diags diag.Diagnostics
	var entries []writeOnlyConfigEntry

	if !config.AWS.IsNull() && !config.AWS.IsUnknown() {
		var aws awsBlockModel
		diags.Append(config.AWS.As(ctx, &aws, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return nil, diags
		}
		if typeutils.IsKnown(aws.ExternalID) {
			entries = append(entries, writeOnlyConfigEntry{
				attributePath: writeOnlyAttributeAWSExternalID,
				value:         aws.ExternalID,
			})
		}
	}

	if !config.Vars.IsNull() && !config.Vars.IsUnknown() {
		var elems map[string]cloudConnectorVarsElement
		diags.Append(config.Vars.ElementsAs(ctx, &elems, false)...)
		if diags.HasError() {
			return nil, diags
		}
		for key, elem := range elems {
			if typeutils.IsKnown(elem.SecretValue) {
				entries = append(entries, writeOnlyConfigEntry{
					attributePath: varsSecretValueAttributePath(key),
					value:         elem.SecretValue,
				})
			}
		}
	}

	return entries, diags
}

func evaluateWriteOnlyDrift(
	ctx context.Context,
	hasher *writeonlyhash.Hasher,
	config cloudConnectorModel,
	priv privateData,
) ([]driftResult, diag.Diagnostics) {
	var diags diag.Diagnostics
	if priv == nil {
		return nil, diags
	}

	entries, entryDiags := writeOnlyEntriesFromConfig(ctx, config)
	diags.Append(entryDiags...)
	if diags.HasError() {
		return nil, diags
	}

	configPaths := make(map[string]struct{}, len(entries))
	for _, entry := range entries {
		configPaths[entry.attributePath] = struct{}{}
	}

	var results []driftResult
	for _, entry := range entries {
		key := hasher.PrivateStateKey(entry.attributePath)
		storedHash, getDiags := priv.GetKey(ctx, key)
		diags.Append(getDiags...)
		if diags.HasError() {
			return nil, diags
		}
		if result := detectWriteOnlyDrift(hasher, entry.attributePath, entry.value, storedHash); result.Changed {
			results = append(results, result)
		}
	}

	indexBytes, indexDiags := priv.GetKey(ctx, varsSecretIndexPrivateStateKey)
	diags.Append(indexDiags...)
	if diags.HasError() {
		return nil, diags
	}
	indexedKeys, indexParseDiags := parseVarsSecretIndex(indexBytes)
	diags.Append(indexParseDiags...)
	if diags.HasError() {
		return nil, diags
	}

	for _, varKey := range indexedKeys {
		attrPath := varsSecretValueAttributePath(varKey)
		if _, inConfig := configPaths[attrPath]; inConfig {
			continue
		}
		key := hasher.PrivateStateKey(attrPath)
		storedHash, getDiags := priv.GetKey(ctx, key)
		diags.Append(getDiags...)
		if diags.HasError() {
			return nil, diags
		}
		if len(storedHash) > 0 {
			results = append(results, driftResult{
				Changed:       true,
				AttributePath: attrPath,
			})
		}
	}

	if _, inConfig := configPaths[writeOnlyAttributeAWSExternalID]; !inConfig {
		storedHash, getDiags := priv.GetKey(ctx, hasher.PrivateStateKey(writeOnlyAttributeAWSExternalID))
		diags.Append(getDiags...)
		if diags.HasError() {
			return nil, diags
		}
		if len(storedHash) > 0 {
			results = append(results, driftResult{
				Changed:       true,
				AttributePath: writeOnlyAttributeAWSExternalID,
			})
		}
	}

	return results, diags
}

func driftWarningDiagnostic(result driftResult) diag.Diagnostic {
	summary := fmt.Sprintf("Detected a change to write-only attribute %s; the resource will be updated.", result.AttributePath)
	detail := "The configured write-only secret value differs from the value last applied, or the attribute was removed from configuration."
	if result.IsImportBaseline {
		detail = "No prior hash exists in private state for this attribute (for example after terraform import). " +
			"The next apply will baseline the hash of the configured value."
	}
	return diag.NewWarningDiagnostic(summary, detail)
}

func parseVarsSecretIndex(data []byte) ([]string, diag.Diagnostics) {
	var diags diag.Diagnostics
	if len(data) == 0 {
		return nil, diags
	}
	var keys []string
	if err := json.Unmarshal(data, &keys); err != nil {
		diags.AddError("Failed to decode write-only vars index from private state", err.Error())
		return nil, diags
	}
	return keys, diags
}

func marshalVarsSecretIndex(keys []string) ([]byte, diag.Diagnostics) {
	var diags diag.Diagnostics
	if len(keys) == 0 {
		return nil, diags
	}
	data, err := json.Marshal(keys)
	if err != nil {
		diags.AddError("Failed to encode write-only vars index for private state", err.Error())
		return nil, diags
	}
	return data, diags
}
