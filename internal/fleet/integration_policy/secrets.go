package integration_policy

import (
	"context"
	"encoding/json"

	fleetapi "github.com/elastic/terraform-provider-elasticstack/generated/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// The secret store is a map of policy secret reference IDs to the
// original value at time of creation. By replacing the ref when
// marshaling the state back to Terraform, we can prevent resource
// drift.
type secretStore map[string]any

// newSecretStore creates a new secretStore from the resource privateData.
func newSecretStore(ctx context.Context, private privateData) (store secretStore, diags diag.Diagnostics) {
	bytes, diags := private.GetKey(ctx, "secrets")
	if diags != nil {
		return
	}
	if bytes == nil {
		bytes = []byte("{}")
	}

	err := json.Unmarshal(bytes, &store)
	if err != nil {
		diags.AddError("could not unmarshal secret store", err.Error())
		return
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

// pruneRefsFromResponse removes any saved secret refs not present in the API response.
func pruneRefsFromResponse(resp *fleetapi.PackagePolicy, secrets secretStore) {
	refs := make(map[string]any)
	for _, r := range utils.Deref(resp.SecretReferences) {
		refs[*r.Id] = nil
	}

	for id := range secrets {
		if _, ok := refs[id]; !ok {
			delete(secrets, id)
		}
	}
}

// handleRespSecrets extracts the wrapped value from each response var, then
// replaces any secret refs with the original value from secrets if available.
func handleRespSecrets(resp *fleetapi.PackagePolicy, secrets secretStore) {
	handleVars := func(vars map[string]any) {
		for key, val := range vars {
			if mval, ok := val.(map[string]any); ok {
				if wrapped, ok := mval["value"]; ok {
					vars[key] = wrapped
					val = wrapped
				} else {
					// Don't keep null (missing) values
					delete(vars, key)
					continue
				}

				if mval, ok := val.(map[string]any); ok {
					if v, ok := mval["isSecretRef"]; ok && v == true {
						refID := mval["id"].(string)
						if original, ok := secrets[refID]; ok {
							vars[key] = original
						}
					}
				}
			}
		}
	}

	handleVars(utils.Deref(resp.Vars))
	for _, input := range resp.Inputs {
		handleVars(utils.Deref(input.Vars))
		for _, _stream := range utils.Deref(input.Streams) {
			stream := _stream.(map[string]any)
			streamVars := stream["vars"].(map[string]any)
			handleVars(streamVars)
		}
	}
}

// handleReqRespSecrets extracts the wrapped value from each response var, then
// maps any secret refs to the original request value.
func handleReqRespSecrets(req fleetapi.PackagePolicyRequest, resp *fleetapi.PackagePolicy, secrets secretStore) {
	handleVars := func(reqVars map[string]any, respVars map[string]any) {
		for key, val := range respVars {
			if mval, ok := val.(map[string]any); ok {
				if wrapped, ok := mval["value"]; ok {
					respVars[key] = wrapped
					val = wrapped
				} else {
					// Don't keep null (missing) values
					delete(respVars, key)
					continue
				}

				if mval, ok := val.(map[string]any); ok {
					if v, ok := mval["isSecretRef"]; ok && v == true {
						refID := mval["id"].(string)
						original := reqVars[key]
						secrets[refID] = original
						respVars[key] = original
					}
				}
			}
		}
	}

	handleVars(utils.Deref(req.Vars), utils.Deref(resp.Vars))
	for inputID, inputReq := range utils.Deref(req.Inputs) {
		inputResp := resp.Inputs[inputID]
		handleVars(utils.Deref(inputReq.Vars), utils.Deref(inputResp.Vars))
		streamsResp := utils.Deref(inputResp.Streams)
		for streamID, streamReq := range utils.Deref(inputReq.Streams) {
			streamResp := streamsResp[streamID].(map[string]any)
			streamRespVars := streamResp["vars"].(map[string]any)
			handleVars(utils.Deref(streamReq.Vars), streamRespVars)
		}
	}
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
