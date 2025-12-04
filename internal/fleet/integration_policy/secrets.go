package integration_policy

import (
	"context"
	"encoding/json"
	"maps"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// The secret store is a map of policy secret reference IDs to the
// original value at time of creation. By replacing the ref when
// marshaling the state back to Terraform, we can prevent resource
// drift.
type secretStore map[string]any

// newSecretStore creates a new secretStore from the resource privateData.
// If the store already exists, it is filtered by any references in the resp policy.
func newSecretStore(ctx context.Context, resp *kbapi.PackagePolicy, private privateData) (store secretStore, diags diag.Diagnostics) {
	bytes, nd := private.GetKey(ctx, "secrets")
	diags.Append(nd...)
	if diags.HasError() {
		return
	}

	if len(bytes) == 0 {
		store = secretStore{}
		return
	}

	err := json.Unmarshal(bytes, &store)
	if err != nil {
		diags.AddError("could not unmarshal secret store", err.Error())
		return
	}

	// Remove any saved secret refs not present in the API response.
	refs := make(map[string]any)
	for _, r := range utils.Deref(resp.SecretReferences) {
		refs[r.Id] = nil
	}

	for id := range store {
		if _, ok := refs[id]; !ok {
			delete(store, id)
		}
	}

	return
}

// Save marshals the secretStore back to the provider.
func (s secretStore) Save(ctx context.Context, private privateData) (diags diag.Diagnostics) {
	bytes, err := json.Marshal(s)
	if err != nil {
		diags.AddError("could not marshal secret store", err.Error())
		return
	}

	return private.SetKey(ctx, "secrets", bytes)
}

// HandleRespSecrets extracts the wrapped value from each response var, then
// replaces any secret refs with the original value from secrets if available.
func HandleRespSecrets(ctx context.Context, resp *kbapi.PackagePolicy, private privateData) (diags diag.Diagnostics) {
	secrets, nd := newSecretStore(ctx, resp, private)
	diags.Append(nd...)
	if diags.HasError() {
		return
	}

	handleVar := func(key string, mval map[string]any, vars map[string]any) {
		if refID, ok := mval["id"]; ok {
			if original, ok := secrets[refID.(string)]; ok {
				vars[key] = original
			}
		} else if ids, ok := mval["ids"]; ok {
			values := []any{}
			for _, id := range ids.([]any) {
				if original, ok := secrets[id.(string)]; ok {
					values = append(values, original)
				}
			}

			if len(values) > 0 {
				vars[key] = values
			}
		}
	}

	handleVars := func(vars map[string]any) {
		for key, val := range vars {
			if mval, ok := val.(map[string]any); ok {
				if wrapped, ok := mval["value"]; ok {
					vars[key] = wrapped
					val = wrapped
				} else if v, ok := mval["isSecretRef"]; ok && v == true {
					handleVar(key, mval, vars)
				} else {
					// Don't keep null (missing) values
					delete(vars, key)
					continue
				}

				if mval, ok := val.(map[string]any); ok {
					if v, ok := mval["isSecretRef"]; ok && v == true {
						handleVar(key, mval, vars)
					}
				}
			}
		}
	}

	handleVars(utils.Deref(resp.Vars))

	// Mapped inputs: extract, mutate, and write back
	mappedInputs, _ := resp.Inputs.AsPackagePolicyMappedInputs()
	typedInputs, _ := resp.Inputs.AsPackagePolicyTypedInputs()
	if mappedInputs != nil {
		for _, input := range mappedInputs {
			if input.Vars != nil {
				handleVars(*input.Vars)
			}

			if input.Streams != nil {
				for _, stream := range *input.Streams {
					if stream.Vars != nil {
						handleVars(*stream.Vars)
					}
				}
			}
		}
		// Write back the mutated data
		if err := resp.Inputs.FromPackagePolicyMappedInputs(mappedInputs); err != nil {
			diags.AddError("could not write back mapped inputs", err.Error())
		}
	} else if typedInputs != nil {
		// Typed inputs: mutate in place through pointers
		for i := range typedInputs {
			if typedInputs[i].Vars != nil {
				handleVars(*typedInputs[i].Vars)
			}
			for j := range typedInputs[i].Streams {
				if typedInputs[i].Streams[j].Vars != nil {
					handleVars(*typedInputs[i].Streams[j].Vars)
				}
			}
		}
		if err := resp.Inputs.FromPackagePolicyTypedInputs(typedInputs); err != nil {
			diags.AddError("could not write back typed inputs", err.Error())
		}
	}

	diags.Append(secrets.Save(ctx, private)...)

	return
}

// HandleReqRespSecrets extracts the wrapped value from each response var, then
// maps any secret refs to the original request value.
func HandleReqRespSecrets(ctx context.Context, req kbapi.PackagePolicyRequest, resp *kbapi.PackagePolicy, private privateData) (diags diag.Diagnostics) {
	secrets, nd := newSecretStore(ctx, resp, private)
	diags.Append(nd...)
	if diags.HasError() {
		return
	}

	handleVar := func(key string, mval map[string]any, reqVars map[string]any, respVars map[string]any) {
		if v, ok := mval["isSecretRef"]; ok && v == true {
			original := reqVars[key]
			respVars[key] = original

			// Is the original also a secret ref?
			// This should only show up during importing and pre 0.11.7 migration.
			if moriginal, ok := original.(map[string]any); ok {
				if v, ok := moriginal["isSecretRef"]; ok && v == true {
					return
				}
			}

			if refID, ok := mval["id"]; ok {
				secrets[refID.(string)] = original
			} else if ids, ok := mval["ids"]; ok {
				originals, ok := original.([]any)
				if !ok || len(originals) != len(ids.([]any)) {
					diags.AddError("mismatched secret ref ids and original values", "the number of secret ref ids does not match the number of original values")
					return
				}

				// Map each id to the corresponding original value by position.
				// The API does not return the original value with the id,
				// so we have to assume the order is preserved.
				for i, id := range ids.([]any) {
					secrets[id.(string)] = originals[i]
				}
			}
		}
	}

	handleVars := func(reqVars map[string]any, respVars map[string]any) {
		for key, val := range respVars {
			if mval, ok := val.(map[string]any); ok {
				if wrapped, ok := mval["value"]; ok {
					respVars[key] = wrapped
					val = wrapped
				} else if v, ok := mval["isSecretRef"]; ok && v == true {
					handleVar(key, mval, reqVars, respVars)
				} else {
					// Don't keep null (missing) values
					delete(respVars, key)
					continue
				}

				if mval, ok := val.(map[string]any); ok {
					handleVar(key, mval, reqVars, respVars)
				}
			}
		}
	}

	mappedReq, _ := req.AsPackagePolicyRequestMappedInputs()
	typedReq, _ := req.AsPackagePolicyRequestTypedInputs()

	reqVars := map[string]any{}
	maps.Copy(reqVars, utils.Deref(mappedReq.Vars))
	maps.Copy(reqVars, utils.Deref(typedReq.Vars))
	// Policy-level vars
	if resp.Vars != nil {
		handleVars(reqVars, utils.Deref(resp.Vars))
	}

	respMappedInputs, _ := resp.Inputs.AsPackagePolicyMappedInputs()
	respTypedInputs, _ := resp.Inputs.AsPackagePolicyTypedInputs()
	if respMappedInputs != nil {
		for inputID, inputReq := range utils.Deref(mappedReq.Inputs) {
			inputResp := respMappedInputs[inputID]
			handleVars(utils.Deref(inputReq.Vars), utils.Deref(inputResp.Vars))

			streamsResp := utils.Deref(inputResp.Streams)
			for streamID, streamReq := range utils.Deref(inputReq.Streams) {
				handleVars(utils.Deref(streamReq.Vars), utils.Deref(streamsResp[streamID].Vars))
			}
		}
		if err := resp.Inputs.FromPackagePolicyMappedInputs(respMappedInputs); err != nil {
			diags.AddError("failed to update mapped inputs", err.Error())
		}
	} else if respTypedInputs != nil {
		for inputIdx, inputReq := range utils.Deref(typedReq.Inputs) {
			inputResp := respTypedInputs[inputIdx]
			handleVars(utils.Deref(inputReq.Vars), utils.Deref(inputResp.Vars))
			for streamIdx, streamReq := range utils.Deref(inputReq.Streams) {
				handleVars(utils.Deref(streamReq.Vars), utils.Deref(inputResp.Streams[streamIdx].Vars))
			}
		}
		if err := resp.Inputs.FromPackagePolicyTypedInputs(respTypedInputs); err != nil {
			diags.AddError("failed to update typed inputs", err.Error())
		}
	}

	diags.Append(secrets.Save(ctx, private)...)

	return
}

// Equivalent to privatestate.ProviderData
type privateData interface {
	// GetKey returns the private state data associated with the given key.
	//
	// If the key is reserved for framework usage, an error diagnostic
	// is returned. If the key is valid, but private state data is not found,
	// nil is returned.
	//
	// The naming of keys only matters in context of a single resource,
	// however care should be taken that any historical keys are not reused
	// without accounting for older resource instances that may still have
	// older data at the key.
	GetKey(ctx context.Context, key string) ([]byte, diag.Diagnostics)

	// SetKey sets the private state data at the given key.
	//
	// If the key is reserved for framework usage, an error diagnostic
	// is returned. The data must be valid JSON and UTF-8 safe or an error
	// diagnostic is returned.
	//
	// The naming of keys only matters in context of a single resource,
	// however care should be taken that any historical keys are not reused
	// without accounting for older resource instances that may still have
	// older data at the key.
	SetKey(ctx context.Context, key string, value []byte) diag.Diagnostics
}
