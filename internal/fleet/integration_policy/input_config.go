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
	"bytes"
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
// policy template, and streams additionally by their data stream dataset. The
// policy revision the snapshot was read at is kept to detect later writes.
type inputConfigSnapshot struct {
	revision float32
	inputs   map[string]map[string]any
	streams  map[string]map[string]any
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

// countIdentities counts every input and stream identity of a list of typed
// inputs, whether or not the element has config. An identity that occurs more
// than once cannot be matched to a single element.
func countIdentities(inputs []map[string]any) map[string]int {
	counts := map[string]int{}
	for _, input := range inputs {
		inputID := inputIdentity(input)
		counts[inputID]++
		streams, _ := input[inputStreamsKey].([]any)
		for _, s := range streams {
			stream, ok := s.(map[string]any)
			if !ok {
				continue
			}
			if streamID := streamIdentity(inputID, stream); streamID != missingIdentityKey {
				counts[streamID]++
			}
		}
	}
	return counts
}

// snapshotInputConfig records the non-empty `config` of every input and stream
// of a package policy read in the typed format. Identities that occur more
// than once (counting elements with and without config) are left out, since
// their config could not be restored to the right element unambiguously.
func snapshotInputConfig(policy *kbapi.PackagePolicy) (inputConfigSnapshot, error) {
	if policy == nil {
		return inputConfigSnapshot{}, nil
	}
	inputs, err := rawTypedInputs(policy)
	if err != nil {
		return inputConfigSnapshot{}, err
	}

	counts := countIdentities(inputs)
	snapshot := inputConfigSnapshot{revision: policy.Revision, inputs: map[string]map[string]any{}, streams: map[string]map[string]any{}}
	for _, input := range inputs {
		inputID := inputIdentity(input)
		if cfg, ok := input[inputConfigKey].(map[string]any); ok && len(cfg) > 0 && counts[inputID] == 1 {
			snapshot.inputs[inputID] = cfg
		}
		streams, _ := input[inputStreamsKey].([]any)
		for _, s := range streams {
			stream, ok := s.(map[string]any)
			if !ok {
				continue
			}
			streamID := streamIdentity(inputID, stream)
			if streamID == missingIdentityKey || counts[streamID] != 1 {
				continue
			}
			if cfg, ok := stream[inputConfigKey].(map[string]any); ok && len(cfg) > 0 {
				snapshot.streams[streamID] = cfg
			}
		}
	}
	return snapshot, nil
}

// deepFill copies into dst every key of src that is absent in dst (even when
// its value in src is null) or null in dst (when its value in src is not),
// recursing when both sides hold objects. Values already present in dst win.
// It reports whether dst was modified.
func deepFill(dst, src map[string]any) bool {
	changed := false
	for key, srcValue := range src {
		dstValue, ok := dst[key]
		if !ok {
			dst[key] = srcValue
			changed = true
			continue
		}
		if dstValue == nil {
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

// typedRequestInputs strips the compiled fields from typed inputs read from the
// API and converts them to the typed update request shape. It also returns the
// JSON encoding of that request shape, which is what would be sent.
func typedRequestInputs(inputs []map[string]any) ([]kbapi.PackagePolicyRequestTypedInput, []byte, error) {
	for _, input := range inputs {
		delete(input, compiledInputKey)
		streams, _ := input[inputStreamsKey].([]any)
		for _, s := range streams {
			if stream, ok := s.(map[string]any); ok {
				delete(stream, compiledStreamKey)
			}
		}
	}
	data, err := json.Marshal(inputs)
	if err != nil {
		return nil, nil, err
	}
	var requestInputs []kbapi.PackagePolicyRequestTypedInput
	if err := json.Unmarshal(data, &requestInputs); err != nil {
		return nil, nil, err
	}
	encoded, err := json.Marshal(requestInputs)
	if err != nil {
		return nil, nil, err
	}
	return requestInputs, encoded, nil
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

	// Only identities that are unique after the update as well are restored,
	// so config never lands on more than one element.
	counts := countIdentities(inputs)
	changed := false
	for _, input := range inputs {
		inputID := inputIdentity(input)
		if counts[inputID] == 1 && fillConfig(input, snapshot.inputs[inputID]) {
			changed = true
		}
		streams, _ := input[inputStreamsKey].([]any)
		for _, s := range streams {
			stream, ok := s.(map[string]any)
			if !ok {
				continue
			}
			streamID := streamIdentity(inputID, stream)
			if streamID != missingIdentityKey && counts[streamID] == 1 && fillConfig(stream, snapshot.streams[streamID]) {
				changed = true
			}
		}
	}
	if !changed {
		return nil, false, nil
	}

	requestInputs, restored, err := typedRequestInputs(inputs)
	if err != nil {
		return nil, false, err
	}

	// The typed request omits some nulls (a config entry's null `value`, `type`
	// or `frozen`), so a restored null can encode to the same request as the
	// current inputs. Nothing would be restored then; send no update.
	current, err := rawTypedInputs(policy)
	if err != nil {
		return nil, false, err
	}
	_, unchanged, err := typedRequestInputs(current)
	if err != nil {
		return nil, false, err
	}
	if bytes.Equal(restored, unchanged) {
		return nil, false, nil
	}
	return requestInputs, true, nil
}

// readInputConfigSnapshot reads the package policy in the typed format and
// returns the config of its inputs and streams. A policy that is not found is
// an error, so that the update is never sent without a snapshot.
func readInputConfigSnapshot(ctx context.Context, client *fleet.Client, policyID, spaceID string) (inputConfigSnapshot, diag.Diagnostics) {
	policy, diags := fleet.GetPackagePolicyTyped(ctx, client, policyID, spaceID)
	if diags.HasError() {
		return inputConfigSnapshot{}, diags
	}
	if policy == nil {
		diags.AddError(
			"Failed to read integration policy input config",
			fmt.Sprintf("The integration policy %q was not found when reading its input config before the update.", policyID),
		)
		return inputConfigSnapshot{}, diags
	}
	snapshot, err := snapshotInputConfig(policy)
	if err != nil {
		diags.AddError("Failed to read integration policy input config", err.Error())
	}
	return snapshot, diags
}

// restorePreservedInputConfig writes back any input or stream config that the
// simplified update removed. The restoring request carries only the version,
// the agent policy assignment and the inputs. It issues no request when the
// snapshot is empty, and no update when nothing was removed. It refuses to
// restore when the policy was written by anyone else since the snapshot.
func restorePreservedInputConfig(ctx context.Context, client *fleet.Client, policyID, spaceID string, snapshot inputConfigSnapshot) diag.Diagnostics {
	if snapshot.isEmpty() {
		return nil
	}

	policy, diags := fleet.GetPackagePolicyTyped(ctx, client, policyID, spaceID)
	if diags.HasError() {
		return diags
	}
	if policy == nil {
		diags.AddError(
			"Failed to restore integration policy input config",
			fmt.Sprintf("The integration policy %q was not found when re-reading it to restore input config after the update.", policyID),
		)
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

	// Kibana bumps the revision by exactly one on every update, and the
	// simplified update ignores the version token, so it cannot be made
	// conditional on the snapshot. Any revision other than the snapshot's plus
	// one means another client wrote the policy since the snapshot, and the
	// snapshot may hold config that was deliberately removed in the meantime
	// (for example an Osquery pack). Restoring it would bring that back.
	if policy.Revision != snapshot.revision+1 {
		diags.AddError(
			"Integration policy modified during update",
			fmt.Sprintf("The integration policy %q was modified by another client while it was being updated "+
				"(revision %d before the update, %d after it, expected %d). The input config read before the update "+
				"may no longer be current, so it was not written back.",
				policyID, int64(snapshot.revision), int64(policy.Revision), int64(snapshot.revision)+1),
		)
		return diags
	}

	// Send the agent policy assignment back as read: Kibana 8.x clears
	// policy_id on a typed update that omits policy_ids.
	req := kbapi.PackagePolicyUpdateRequestTypedInputs{
		Inputs:    &inputs,
		Version:   policy.Version,
		PolicyId:  policy.PolicyId,
		PolicyIds: policy.PolicyIds,
	}
	_, updateDiags := fleet.UpdatePackagePolicyTyped(ctx, client, policyID, spaceID, req)
	diags.Append(updateDiags...)
	return diags
}

// updatePackagePolicyPreservingInputConfig sends the simplified update and
// writes back any input or stream config it removed. The config is
// snapshotted first; when that read fails no update is sent, since the update
// would otherwise drop the config without any way to restore it.
func updatePackagePolicyPreservingInputConfig(
	ctx context.Context, client *fleet.Client, policyID, spaceID string, body kbapi.PackagePolicyRequestMappedInputs,
) (*kbapi.PackagePolicy, diag.Diagnostics) {
	snapshot, diags := readInputConfigSnapshot(ctx, client, policyID, spaceID)
	if diags.HasError() {
		return nil, diags
	}

	policy, updateDiags := fleet.UpdatePackagePolicy(ctx, client, policyID, spaceID, body)
	diags.Append(updateDiags...)
	if diags.HasError() {
		return nil, diags
	}

	restoreDiags := restorePreservedInputConfig(ctx, client, policyID, spaceID, snapshot)
	diags.Append(restoreDiags...)
	if restoreDiags.HasError() {
		// The simplified update has been applied at this point, so any config it
		// removed is gone from the server, and the next update's snapshot will
		// not contain it either.
		diags.AddError(
			"Integration policy input config not restored",
			fmt.Sprintf("The integration policy %q was updated, but any input or stream config the update removed "+
				"could not be written back (see the other errors). Re-running apply does not restore it, because the "+
				"next update reads the policy without that config. Check the input config of the integration policy "+
				"in Kibana, for example APM API keys or Osquery packs, and recreate it where needed.", policyID),
		)
		return nil, diags
	}
	return policy, diags
}
