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
func reconcileManagedIntegrationSecretsFromPrior(ctx context.Context, prior, populated *agentlessPolicyModel, diags *diag.Diagnostics) {
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

func packageNameVersionFromModel(ctx context.Context, m *agentlessPolicyModel, diags *diag.Diagnostics) (name, version string) {
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
	priorMap, ok := varsJSONToMap(prior, attrPath, diags)
	if !ok || len(priorMap) == 0 {
		return populated
	}
	respMap, ok := varsJSONToMap(populated, attrPath, diags)
	if !ok {
		return populated
	}
	reconcileSecretVarsMapFromPrior(priorMap, respMap)
	return varsJSONFromMap(ctx, respMap, packageName, packageVersion, attrPath, diags)
}

func varsJSONToMap(v policyshape.VarsJSONValue, attrPath path.Path, diags *diag.Diagnostics) (map[string]any, bool) {
	if v.IsNull() || v.IsUnknown() {
		return nil, false
	}
	raw := v.ValueString()
	if raw == "" {
		return nil, false
	}
	var out map[string]any
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		diags.AddAttributeError(attrPath, "Failed to decode vars_json for secret reconciliation", err.Error())
		return nil, false
	}
	return out, true
}

func varsJSONFromMap(_ context.Context, vars map[string]any, packageName, packageVersion string, attrPath path.Path, diags *diag.Diagnostics) policyshape.VarsJSONValue {
	if len(vars) == 0 {
		return policyshape.NewVarsJSONNull()
	}
	b, err := json.Marshal(vars)
	if err != nil {
		diags.AddAttributeError(attrPath, "Failed to marshal vars_json after secret reconciliation", err.Error())
		return policyshape.NewVarsJSONNull()
	}
	v, d := policyshape.NewVarsJSONWithIntegration(string(b), packageName, packageVersion, lookupCachedPackageInfo)
	diags.Append(d...)
	return v
}

func reconcileInputsSecretsFromPrior(ctx context.Context, prior, populated *agentlessPolicyModel, diags *diag.Diagnostics) {
	priorInputs := typeutils.MapTypeAs[agentlessInputModel](ctx, prior.Inputs.MapValue, path.Root(attrInputs), diags)
	popInputs := typeutils.MapTypeAs[agentlessInputModel](ctx, populated.Inputs.MapValue, path.Root(attrInputs), diags)
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

	inputsValue, d := policyshape.NewInputsValueFrom(ctx, agentlessInputType(), popInputs)
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
	priorMap, ok := normalizedVarsToMap(prior, attrPath, diags)
	if !ok || len(priorMap) == 0 {
		return populated
	}
	respMap, ok := normalizedVarsToMap(populated, attrPath, diags)
	if !ok {
		return populated
	}
	reconcileSecretVarsMapFromPrior(priorMap, respMap)
	return normalizedVarsFromMap(respMap, attrPath, diags)
}

func normalizedVarsToMap(n jsontypes.Normalized, attrPath path.Path, diags *diag.Diagnostics) (map[string]any, bool) {
	if n.IsNull() || n.IsUnknown() {
		return nil, false
	}
	raw := n.ValueString()
	if raw == "" {
		return nil, false
	}
	var out map[string]any
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		diags.AddAttributeError(attrPath, "Failed to decode vars for secret reconciliation", err.Error())
		return nil, false
	}
	return out, true
}

func normalizedVarsFromMap(vars map[string]any, attrPath path.Path, diags *diag.Diagnostics) jsontypes.Normalized {
	if len(vars) == 0 {
		return jsontypes.NewNormalizedNull()
	}
	b, err := json.Marshal(vars)
	if err != nil {
		diags.AddAttributeError(attrPath, "Failed to marshal vars after secret reconciliation", err.Error())
		return jsontypes.NewNormalizedNull()
	}
	return jsontypes.NewNormalizedValue(string(b))
}

// reconcileSecretVarsMapFromPrior mutates resp in place, mirroring
// policyshape.HandleReqRespSecrets but using prior config/state as the source
// of plaintext instead of a private secret store.
func reconcileSecretVarsMapFromPrior(prior, resp map[string]any) {
	if prior == nil || resp == nil {
		return
	}
	for key, val := range resp {
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
				if priorVal, ok := prior[key]; ok {
					resp[key] = priorVal
				}
			}
		}
	}
}

func isSecretRefMap(m map[string]any) bool {
	if v, ok := m["isSecretRef"]; ok && v == true {
		return true
	}
	return false
}
