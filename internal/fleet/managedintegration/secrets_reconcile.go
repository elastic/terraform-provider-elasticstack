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
	"context"
	"encoding/json"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/fleet/policyshape"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// reconcileManagedIntegrationSecretsFromPrior replaces Fleet secret-reference
// shapes echoed by GET with the practitioner's prior plaintext (or prior plan)
// so read-after-write and refresh do not produce inconsistent results. When
// prior has no configured value (import without config), API refs are kept.
func reconcileManagedIntegrationSecretsFromPrior(ctx context.Context, prior, populated *managedIntegrationModel, diags *diag.Diagnostics) {
	if prior == nil || populated == nil {
		return
	}

	pkgName, pkgVersion := packageNameVersionFromModel(ctx, populated, diags)
	if diags.HasError() {
		return
	}

	if typeutils.IsKnown(prior.VarsJSON) && typeutils.IsKnown(populated.VarsJSON) {
		populated.VarsJSON = reconcileVarsJSONFromPrior(ctx, prior.VarsJSON, populated.VarsJSON, pkgName, pkgVersion, path.Root(attrVarsJSON), diags)
	}

	if typeutils.IsKnown(prior.Inputs.MapValue) && typeutils.IsKnown(populated.Inputs.MapValue) {
		reconcileInputsSecretsFromPrior(ctx, prior, populated, diags)
	}
}

func packageNameVersionFromModel(ctx context.Context, m *managedIntegrationModel, diags *diag.Diagnostics) (name, version string) {
	if !typeutils.IsKnown(m.Package) || m.Package.IsNull() {
		return "", ""
	}
	var pkg packageModel
	diags.Append(m.Package.As(ctx, &pkg, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return "", ""
	}
	return pkg.Name.ValueString(), pkg.Version.ValueString()
}

func reconcileVarsJSONFromPrior(
	ctx context.Context,
	prior, populated policyshape.VarsJSONValue,
	packageName, packageVersion string,
	attrPath path.Path,
	diags *diag.Diagnostics,
) policyshape.VarsJSONValue {
	if populated.IsNull() {
		return populated
	}
	respMap, ok := reconcileVarsMapFromJSON(prior, populated, secretReconcileAttrVarsJSON, attrPath, diags)
	if !ok {
		return populated
	}
	return varsJSONFromMap(ctx, respMap, packageName, packageVersion, attrPath, diags)
}

const (
	secretReconcileAttrVarsJSON = "vars_json"
	secretReconcileAttrVars     = "vars"
)

type secretReconcileJSONString interface {
	IsNull() bool
	IsUnknown() bool
	ValueString() string
}

func secretReconcileJSONToMap(v secretReconcileJSONString, attrPath path.Path, attrLabel string, diags *diag.Diagnostics) (map[string]any, bool) {
	if v.IsNull() || v.IsUnknown() {
		return nil, false
	}
	raw := v.ValueString()
	if raw == "" {
		return nil, false
	}
	var out map[string]any
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		diags.AddAttributeError(attrPath, fmt.Sprintf("Failed to decode %s for secret reconciliation", attrLabel), err.Error())
		return nil, false
	}
	return out, true
}

func marshalSecretReconcileVarsMap(vars map[string]any, attrPath path.Path, attrLabel string, diags *diag.Diagnostics) (raw string, useNull bool) {
	if len(vars) == 0 {
		return "", true
	}
	b, err := json.Marshal(vars)
	if err != nil {
		diags.AddAttributeError(attrPath, fmt.Sprintf("Failed to marshal %s after secret reconciliation", attrLabel), err.Error())
		return "", true
	}
	return string(b), false
}

func varsJSONFromMap(_ context.Context, vars map[string]any, packageName, packageVersion string, attrPath path.Path, diags *diag.Diagnostics) policyshape.VarsJSONValue {
	raw, useNull := marshalSecretReconcileVarsMap(vars, attrPath, secretReconcileAttrVarsJSON, diags)
	if useNull {
		return policyshape.NewVarsJSONNull()
	}
	v, d := policyshape.NewVarsJSONWithIntegration(raw, packageName, packageVersion, lookupCachedPackageInfo)
	diags.Append(d...)
	return v
}

func reconcileInputsSecretsFromPrior(ctx context.Context, prior, populated *managedIntegrationModel, diags *diag.Diagnostics) {
	priorInputs := typeutils.MapTypeAs[managedIntegrationInputModel](ctx, prior.Inputs.MapValue, path.Root(attrInputs), diags)
	popInputs := typeutils.MapTypeAs[managedIntegrationInputModel](ctx, populated.Inputs.MapValue, path.Root(attrInputs), diags)
	if priorInputs == nil || popInputs == nil {
		return
	}

	for inputID, popIn := range popInputs {
		priorIn, ok := priorInputs[inputID]
		if !ok {
			continue
		}
		inputPath := path.Root(attrInputs).AtMapKey(inputID)
		popIn.Vars = reconcileNormalizedVarsFromPrior(priorIn.Vars, popIn.Vars, inputPath.AtName("vars"), diags)

		if typeutils.IsKnown(priorIn.Streams) && typeutils.IsKnown(popIn.Streams) {
			priorStreams := typeutils.MapTypeAs[policyshape.InputStreamModel](ctx, priorIn.Streams, inputPath.AtName("streams"), diags)
			popStreams := typeutils.MapTypeAs[policyshape.InputStreamModel](ctx, popIn.Streams, inputPath.AtName("streams"), diags)
			if priorStreams != nil && popStreams != nil {
				for streamID, popStream := range popStreams {
					if priorStream, ok := priorStreams[streamID]; ok {
						streamPath := inputPath.AtName("streams").AtMapKey(streamID)
						popStream.Vars = reconcileNormalizedVarsFromPrior(priorStream.Vars, popStream.Vars, streamPath.AtName("vars"), diags)
						popStreams[streamID] = popStream
					}
				}
				streamsMap, d := types.MapValueFrom(ctx, policyshape.StreamType(), popStreams)
				diags.Append(d...)
				popIn.Streams = streamsMap
			}
		}
		popInputs[inputID] = popIn
	}

	inputsValue, d := policyshape.NewInputsValueFrom(ctx, managedIntegrationInputType(), popInputs)
	diags.Append(d...)
	populated.Inputs = inputsValue
}

func reconcileNormalizedVarsFromPrior(prior, populated jsontypes.Normalized, attrPath path.Path, diags *diag.Diagnostics) jsontypes.Normalized {
	if populated.IsNull() || populated.IsUnknown() {
		return populated
	}
	if prior.IsNull() || prior.IsUnknown() {
		return populated
	}
	respMap, ok := reconcileVarsMapFromJSON(prior, populated, secretReconcileAttrVars, attrPath, diags)
	if !ok {
		return populated
	}
	return normalizedVarsFromMap(respMap, attrPath, diags)
}

func reconcileVarsMapFromJSON(prior, populated secretReconcileJSONString, attrLabel string, attrPath path.Path, diags *diag.Diagnostics) (map[string]any, bool) {
	priorMap, ok := secretReconcileJSONToMap(prior, attrPath, attrLabel, diags)
	if !ok || len(priorMap) == 0 {
		return nil, false
	}
	respMap, ok := secretReconcileJSONToMap(populated, attrPath, attrLabel, diags)
	if !ok {
		return nil, false
	}
	reconcileSecretVarsMapFromPrior(priorMap, respMap, attrPath, diags)
	return respMap, true
}

func normalizedVarsFromMap(vars map[string]any, attrPath path.Path, diags *diag.Diagnostics) jsontypes.Normalized {
	raw, useNull := marshalSecretReconcileVarsMap(vars, attrPath, secretReconcileAttrVars, diags)
	if useNull {
		return jsontypes.NewNormalizedNull()
	}
	return jsontypes.NewNormalizedValue(raw)
}

// reconcileSecretVarsMapFromPrior mutates resp in place, mirroring
// policyshape.HandleReqRespSecrets but using prior config/state as the source
// of plaintext instead of a private secret store.
func reconcileSecretVarsMapFromPrior(prior, resp map[string]any, basePath path.Path, diags *diag.Diagnostics) {
	if prior == nil || resp == nil {
		return
	}
	for key, val := range resp {
		attrPath := basePath.AtMapKey(key)
		mval, ok := val.(map[string]any)
		if !ok {
			continue
		}
		if wrapped, ok := mval["value"]; ok {
			resp[key] = wrapped
			val = wrapped
		}
		if mval, ok := val.(map[string]any); ok {
			if isSecretRefMap(mval) {
				reconcileSecretRefValueFromPrior(key, mval, prior, resp, attrPath, diags)
			}
		}
	}
}

func reconcileSecretRefValueFromPrior(key string, mval map[string]any, prior, resp map[string]any, attrPath path.Path, diags *diag.Diagnostics) {
	priorVal, ok := prior[key]
	if !ok {
		return
	}
	if ids, ok := mval["ids"]; ok {
		idSlice, ok := coerceAnySlice(ids)
		if !ok {
			diags.AddAttributeError(attrPath, "Failed to reconcile Fleet secret reference", fmt.Sprintf("unexpected secret reference ids type %T", ids))
			return
		}
		originals, ok := coerceAnySlice(priorVal)
		if !ok {
			diags.AddAttributeError(attrPath, "Failed to reconcile Fleet secret reference", fmt.Sprintf("prior value is not a list (got %T)", priorVal))
			return
		}
		if len(originals) != len(idSlice) {
			diags.AddAttributeError(attrPath, "Failed to reconcile Fleet secret reference",
				fmt.Sprintf("secret reference id count (%d) does not match configured prior value count (%d)", len(idSlice), len(originals)))
			return
		}
		resp[key] = priorVal
		return
	}
	if priorRef, ok := priorVal.(map[string]any); ok {
		if isSecretRefMap(priorRef) {
			resp[key] = priorVal
			return
		}
	}
	resp[key] = priorVal
}

// coerceAnySlice normalizes a decoded-JSON list value into []any, accepting
// either the []any that json.Unmarshal produces or a []string a caller may
// have supplied directly. The bool reports whether v was a list at all.
func coerceAnySlice(v any) ([]any, bool) {
	switch x := v.(type) {
	case []any:
		return x, true
	case []string:
		out := make([]any, len(x))
		for i, s := range x {
			out[i] = s
		}
		return out, true
	default:
		return nil, false
	}
}

func isSecretRefMap(m map[string]any) bool {
	return m["isSecretRef"] == true
}
