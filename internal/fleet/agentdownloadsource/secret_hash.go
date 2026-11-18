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

package agentdownloadsource

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/writeonlyhash"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// secretHashKey is the private-state key for the hash of the write-only
// attribute at attrPath.
func secretHashKey(attrPath string) string {
	return secretHasher.PrivateStateKey(attrPath)
}

// loadStoredHashes reads the stored hash of every write-only attribute, keyed
// by attribute path. Attributes without a stored hash are absent from the map.
func loadStoredHashes(ctx context.Context, private entitycore.PrivateStateStorage) (map[string][]byte, diag.Diagnostics) {
	var diags diag.Diagnostics
	hashes := map[string][]byte{}
	if private == nil {
		return hashes, diags
	}
	for _, attrPath := range writeOnlyPaths {
		raw, getDiags := private.GetKey(ctx, secretHashKey(attrPath))
		diags.Append(getDiags...)
		if diags.HasError() {
			return nil, diags
		}
		hash, err := secretHasher.DecodeFromPrivateState(raw)
		if err != nil {
			diags.AddError("Failed to decode write-only secret hash", err.Error())
			return nil, diags
		}
		if len(hash) > 0 {
			hashes[attrPath] = hash
		}
	}
	return hashes, diags
}

// persistSecretHashes records the hash of each applied write-only value and
// clears the hash of any write-only attribute that has a stored hash but is no
// longer configured. It must only be called after the remote write succeeded,
// so a failed clearing update leaves the hash in place for the next plan.
func persistSecretHashes(
	ctx context.Context,
	private entitycore.PrivateStateStorage,
	configured map[string]string,
	stored map[string][]byte,
	diags *diag.Diagnostics,
) {
	if private == nil {
		if len(configured) > 0 {
			writeonlyhash.AddPrivateStateUnavailableError(diags)
		}
		return
	}

	for _, attrPath := range writeOnlyPaths {
		value, isConfigured := configured[attrPath]
		if !isConfigured {
			if len(stored[attrPath]) > 0 {
				diags.Append(private.SetKey(ctx, secretHashKey(attrPath), nil)...)
			}
			continue
		}

		if secretHasher.IsUpToDate(value, stored[attrPath]) {
			continue
		}

		encoded, hashDiags := secretHasher.ComputeForPrivateState(value)
		diags.Append(hashDiags...)
		if diags.HasError() {
			return
		}
		diags.Append(private.SetKey(ctx, secretHashKey(attrPath), encoded)...)
		if diags.HasError() {
			return
		}
	}
}

// clearAllSecretHashes removes every write-only hash from private state.
func clearAllSecretHashes(ctx context.Context, private entitycore.PrivateStateStorage, diags *diag.Diagnostics) {
	if private == nil {
		return
	}
	for _, attrPath := range writeOnlyPaths {
		diags.Append(private.SetKey(ctx, secretHashKey(attrPath), nil)...)
	}
}
