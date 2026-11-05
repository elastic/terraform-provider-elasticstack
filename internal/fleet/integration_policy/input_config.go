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

package integrationpolicy

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// The simplified (mapped-inputs) package policy API used by Update only
// carries enabled/vars/streams/condition per input. Kibana rebuilds every input
// from the package template and drops the input- and stream-level `config`
// objects, which are written by Kibana itself or by other Kibana features (for
// example the APM plugin's source map and agent configuration API keys, or
// Osquery packs) and are not managed by this resource. The helpers below
// snapshot that `config` before the update and write back whatever the update
// removed.

const (
	inputConfigKey     = "config"
	inputStreamsKey    = "streams"
	compiledInputKey   = "compiled_input"
	compiledStreamKey  = "compiled_stream"
	inputTypeKey       = "type"
	policyTemplateKey  = "policy_template"
	streamDataStream   = "data_stream"
	streamDatasetKey   = "dataset"
	identitySeparator  = "\x00"
	missingIdentityKey = ""
)

// inputConfigSnapshot holds the `config` objects of a package policy's inputs
// and streams, keyed by a stable identity. Input ids are not stable across the
// typed and simplified representations, so inputs are identified by type and
// policy template, and streams additionally by their data stream dataset.
type inputConfigSnapshot struct {
	inputs  map[string]map[string]any
	streams map[string]map[string]any
}

func (s inputConfigSnapshot) isEmpty() bool {
	return len(s.inputs) == 0 && len(s.streams) == 0
}

func inputIdentity(raw map[string]any) string {
	inputType, _ := raw[inputTypeKey].(string)
	policyTemplate, _ := raw[policyTemplateKey].(string)
	return inputType + identitySeparator + policyTemplate
}

func streamIdentity(inputID string, raw map[string]any) string {
	dataStream, _ := raw[streamDataStream].(map[string]any)
	dataset, _ := dataStream[streamDatasetKey].(string)
	if dataset == "" {
		return missingIdentityKey
	}
	return inputID + identitySeparator + dataset
}

// rawTypedInputs decodes the inputs of a package policy read in the typed
// (non-simplified) format into generic maps, so that no field is lost.
func rawTypedInputs(policy *kbapi.PackagePolicy) ([]map[string]any, error) {
	if policy == nil {
		return nil, nil
	}
	data, err := json.Marshal(policy.Inputs)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 || string(data) == "null" {
		return nil, nil
	}
	var inputs []map[string]any
	if err := json.Unmarshal(data, &inputs); err != nil {
		return nil, fmt.Errorf("package policy inputs are not in the typed format: %w", err)
	}
	return inputs, nil
}

// snapshotInputConfig records the non-empty `config` of every input and stream
// of a package policy read in the typed format. Identities that occur more
// than once are left out, since their config could not be restored to the
// right element unambiguously.
func snapshotInputConfig(policy *kbapi.PackagePolicy) (inputConfigSnapshot, error) {
	inputs, err := rawTypedInputs(policy)
	if err != nil {
		return inputConfigSnapshot{}, err
	}

	snapshot := inputConfigSnapshot{inputs: map[string]map[string]any{}, streams: map[string]map[string]any{}}
	ambiguous := map[string]bool{}
	record := func(target map[string]map[string]any, id string, cfg map[string]any) {
		if ambiguous[id] {
			return
		}
		if _, seen := target[id]; seen {
			delete(target, id)
			ambiguous[id] = true
			return
		}
		target[id] = cfg
	}

	for _, input := range inputs {
		inputID := inputIdentity(input)
		if cfg, ok := input[inputConfigKey].(map[string]any); ok && len(cfg) > 0 {
			record(snapshot.inputs, inputID, cfg)
		}
		streams, _ := input[inputStreamsKey].([]any)
		for _, s := range streams {
			stream, ok := s.(map[string]any)
			if !ok {
				continue
			}
			streamID := streamIdentity(inputID, stream)
			if streamID == missingIdentityKey {
				continue
			}
			if cfg, ok := stream[inputConfigKey].(map[string]any); ok && len(cfg) > 0 {
				record(snapshot.streams, streamID, cfg)
			}
		}
	}
	return snapshot, nil
}

// deepFill copies into dst every key of src that is absent (or null) in dst,
// recursing when both sides hold objects. Values already present in dst win.
// It reports whether dst was modified.
func deepFill(dst, src map[string]any) bool {
	changed := false
	for key, srcValue := range src {
		dstValue, ok := dst[key]
		if !ok || dstValue == nil {
			if srcValue != nil {
				dst[key] = srcValue
				changed = true
			}
			continue
		}
		dstMap, dstIsMap := dstValue.(map[string]any)
		srcMap, srcIsMap := srcValue.(map[string]any)
		if dstIsMap && srcIsMap && deepFill(dstMap, srcMap) {
			changed = true
		}
	}
	return changed
}

// fillConfig restores the snapshot config into the element's `config` object
// and reports whether anything was added.
func fillConfig(element map[string]any, prior map[string]any) bool {
	if len(prior) == 0 {
		return false
	}
	current, ok := element[inputConfigKey].(map[string]any)
	if !ok || current == nil {
		current = map[string]any{}
	}
	if !deepFill(current, prior) {
		return false
	}
	element[inputConfigKey] = current
	return true
}

// restoreInputConfig merges the snapshot into the inputs of a package policy
// read in the typed format after an update. It returns the inputs to send in a
// typed update request (compiled fields removed) and whether any config was
// restored. When nothing was restored the returned inputs are nil.
func restoreInputConfig(snapshot inputConfigSnapshot, policy *kbapi.PackagePolicy) ([]kbapi.PackagePolicyRequestTypedInput, bool, error) {
	if snapshot.isEmpty() {
		return nil, false, nil
	}
	inputs, err := rawTypedInputs(policy)
	if err != nil {
		return nil, false, err
	}

	changed := false
	for _, input := range inputs {
		inputID := inputIdentity(input)
		if fillConfig(input, snapshot.inputs[inputID]) {
			changed = true
		}
		delete(input, compiledInputKey)
		streams, _ := input[inputStreamsKey].([]any)
		for _, s := range streams {
			stream, ok := s.(map[string]any)
			if !ok {
				continue
			}
			delete(stream, compiledStreamKey)
			if streamID := streamIdentity(inputID, stream); streamID != missingIdentityKey && fillConfig(stream, snapshot.streams[streamID]) {
				changed = true
			}
		}
	}
	if !changed {
		return nil, false, nil
	}

	data, err := json.Marshal(inputs)
	if err != nil {
		return nil, false, err
	}
	var requestInputs []kbapi.PackagePolicyRequestTypedInput
	if err := json.Unmarshal(data, &requestInputs); err != nil {
		return nil, false, err
	}
	return requestInputs, true, nil
}

// readInputConfigSnapshot reads the package policy in the typed format and
// returns the config of its inputs and streams.
func readInputConfigSnapshot(ctx context.Context, client *fleet.Client, policyID, spaceID string) (inputConfigSnapshot, diag.Diagnostics) {
	policy, diags := fleet.GetPackagePolicyTyped(ctx, client, policyID, spaceID)
	if diags.HasError() {
		return inputConfigSnapshot{}, diags
	}
	if policy == nil {
		return inputConfigSnapshot{}, diags
	}
	snapshot, err := snapshotInputConfig(policy)
	if err != nil {
		diags.AddError("Failed to read integration policy input config", err.Error())
	}
	return snapshot, diags
}

// restorePreservedInputConfig writes back any input or stream config that the
// simplified update removed. It issues no request when the snapshot is empty,
// and no update when nothing was removed.
func restorePreservedInputConfig(ctx context.Context, client *fleet.Client, policyID, spaceID string, snapshot inputConfigSnapshot) diag.Diagnostics {
	if snapshot.isEmpty() {
		return nil
	}

	policy, diags := fleet.GetPackagePolicyTyped(ctx, client, policyID, spaceID)
	if diags.HasError() || policy == nil {
		return diags
	}

	inputs, changed, err := restoreInputConfig(snapshot, policy)
	if err != nil {
		diags.AddError("Failed to restore integration policy input config", err.Error())
		return diags
	}
	if !changed {
		return diags
	}

	req := kbapi.PackagePolicyUpdateRequestTypedInputs{
		Inputs:  &inputs,
		Version: policy.Version,
	}
	_, updateDiags := fleet.UpdatePackagePolicyTyped(ctx, client, policyID, spaceID, req)
	diags.Append(updateDiags...)
	return diags
}
