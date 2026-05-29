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

// Package writeonlyhash provides bcrypt-backed hashing for write-only secret
// attributes stored in Terraform resource private state.
//
// Resources use these hashes during ModifyPlan to detect when a write-only
// secret value in configuration differs from the value last applied, without
// ever persisting the raw secret in state.
//
// # Domain separation, SHA-256 pre-hash, and per-resource-type salt
//
// bcrypt.GenerateFromPassword generates a fresh random salt on each call and
// embeds it in the returned hash. To ensure the same secret value produces
// different hashes across resource types (rainbow-table protection across state
// files), the helper domain-separates the input before hashing:
//
//	SHA-256(resourceTypeName + "\x00" + value)
//
// The 32-byte digest is then passed to bcrypt. bcrypt limits raw password input
// to 72 bytes; with a typical resource-type prefix the effective secret limit
// would be roughly 38 bytes. Pre-hashing removes that ceiling so long secrets
// (JWT-style tokens, base64 access keys, long external IDs, and so on) are
// supported without caller-side length contracts, while preserving domain
// separation and bcrypt's adaptive cost and per-call salt.
//
// The resourceTypeName passed to New is typically the Terraform resource type
// string (for example "elasticstack_fleet_cloud_connector"). An empty
// resourceTypeName is allowed but provides no cross-type separation.
//
// # Private state keys
//
// Use PrivateStateKey to derive stable keys for storing hashes in private
// state, for example "secret_hash:aws.external_id".
package writeonlyhash

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const (
	privateStateKeyPrefix = "secret_hash:"
	domainSeparator       = "\x00"
)

// Hasher computes and verifies bcrypt hashes for write-only secret values.
// Create one Hasher per resource type via New and reuse it for all write-only
// attributes on that resource.
type Hasher struct {
	// Cost is the bcrypt work factor used by Compute. When zero, bcrypt.DefaultCost (10) is used.
	// Set Cost once before the first Compute call on this Hasher; do not mutate it concurrently.
	Cost int

	resourceTypeName string
}

// New returns a Hasher bound to resourceTypeName for domain-separated hashing.
// The name should be stable for the lifetime of the resource type (typically
// the Terraform resource type string). Empty resourceTypeName is allowed.
func New(resourceTypeName string) *Hasher {
	return &Hasher{
		Cost:             bcrypt.DefaultCost,
		resourceTypeName: resourceTypeName,
	}
}

// Compute returns a bcrypt hash of value suitable for storage in resource private state.
// Errors describe the failure without including any part of value.
func (h *Hasher) Compute(value string) ([]byte, error) {
	cost := h.Cost
	if cost == 0 {
		cost = bcrypt.DefaultCost
	}

	hash, err := bcrypt.GenerateFromPassword(h.hashInput(value), cost)
	if err != nil {
		return nil, computeError(err)
	}

	// Terraform resource private state values must be valid JSON.
	return json.Marshal(string(hash))
}

// DecodeStoredHash returns the raw bcrypt hash bytes from a private-state value.
// Values produced by Compute are JSON-encoded strings; legacy raw bcrypt bytes
// are returned unchanged.
func DecodeStoredHash(storedHash []byte) []byte {
	if len(storedHash) == 0 {
		return nil
	}

	var decoded string
	if err := json.Unmarshal(storedHash, &decoded); err == nil {
		return []byte(decoded)
	}

	return storedHash
}

// Matches reports whether value corresponds to storedHash on this Hasher.
// Nil or empty storedHash returns false without error.
func (h *Hasher) Matches(value string, storedHash []byte) bool {
	storedHash = DecodeStoredHash(storedHash)
	if len(storedHash) == 0 {
		return false
	}

	return bcrypt.CompareHashAndPassword(storedHash, h.hashInput(value)) == nil
}

// PrivateStateKey returns a stable private-state key for attributePath.
// The format is "secret_hash:" followed by attributePath verbatim, for example
// PrivateStateKey("aws.external_id") returns "secret_hash:aws.external_id".
// The same attributePath always yields the same key regardless of Hasher instance.
func (h *Hasher) PrivateStateKey(attributePath string) string {
	return privateStateKeyPrefix + attributePath
}

func (h *Hasher) hashInput(value string) []byte {
	domainSeparated := make([]byte, 0, len(h.resourceTypeName)+len(domainSeparator)+len(value))
	domainSeparated = append(domainSeparated, h.resourceTypeName...)
	domainSeparated = append(domainSeparated, domainSeparator...)
	domainSeparated = append(domainSeparated, value...)

	digest := sha256.Sum256(domainSeparated)
	return digest[:]
}

func computeError(err error) error {
	if errors.As(err, new(bcrypt.InvalidCostError)) {
		return fmt.Errorf("bcrypt cost out of range")
	}

	return fmt.Errorf("bcrypt hash computation failed")
}
