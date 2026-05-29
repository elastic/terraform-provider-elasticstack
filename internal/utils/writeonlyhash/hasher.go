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
package writeonlyhash

import (
	"errors"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

const (
	defaultCost = 10
	keyPrefix   = "secret_hash:"
)

// Hasher manages bcrypt-based private-state hashes for write-only attributes.
// It binds a per-resource-type salt so the same secret value produces different
// hashes across resource types. Cost controls bcrypt work factor; when Cost is
// zero, Compute uses the default of 10. A Hasher must not be shared across
// resource types—construct one per resource via New with that resource's
// stable type identifier.
type Hasher struct {
	// Salt is the per-resource-type salt derived from the resourceTypeName
	// passed to New. It namespaces values before bcrypt hashing.
	Salt []byte

	// Cost is the bcrypt cost parameter. Zero means defaultCost (10).
	Cost int
}

// New returns a Hasher whose salt is derived from resourceTypeName. Use a
// stable, unique string per Terraform resource type (for example
// "elasticsearch_connector" or "fleet_cloud_connector"). The same plaintext
// value hashed under different resource type names will not verify on another
// Hasher.
func New(resourceTypeName string) *Hasher {
	return &Hasher{
		Salt: []byte(resourceTypeName),
		Cost: 0,
	}
}

// Compute returns a bcrypt hash of value suitable for storage in resource
// private state. When Cost is zero, the default bcrypt cost of 10 is used.
// Errors never include the input value.
func (h *Hasher) Compute(value string) ([]byte, error) {
	hash, err := bcrypt.GenerateFromPassword(h.passwordBytes(value), h.effectiveCost())
	if err != nil {
		return nil, hashError(err)
	}
	return hash, nil
}

// Matches reports whether value corresponds to storedHash as produced by
// Compute on this Hasher. It returns false when the value does not match or
// when bcrypt comparison fails for any reason.
func (h *Hasher) Matches(value string, storedHash []byte) bool {
	err := bcrypt.CompareHashAndPassword(storedHash, h.passwordBytes(value))
	return err == nil
}

// PrivateStateKey returns a stable private-state key for the write-only
// attribute identified by attributePath. The format is
// secret_hash:<attributePath> (for example secret_hash:aws.external_id or
// secret_hash:configuration_values["password"].secret_value). attributePath is
// not modified; the caller must supply a path that uniquely identifies the
// attribute, including map key indices in bracket notation.
func (h *Hasher) PrivateStateKey(attributePath string) string {
	return keyPrefix + attributePath
}

func (h *Hasher) effectiveCost() int {
	if h.Cost == 0 {
		return defaultCost
	}
	return h.Cost
}

func (h *Hasher) passwordBytes(value string) []byte {
	prefix := make([]byte, 0, len(h.Salt)+1+len(value))
	prefix = append(prefix, h.Salt...)
	prefix = append(prefix, ':')
	prefix = append(prefix, value...)
	return prefix
}

func hashError(err error) error {
	if strings.Contains(err.Error(), "cost") {
		return errors.New("bcrypt cost out of range")
	}
	return errors.New("failed to hash value")
}
