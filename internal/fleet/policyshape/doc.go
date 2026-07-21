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

// Package policyshape provides the shared Plugin Framework modeling for
// Fleet package policy inputs, streams, and vars: the "shape" of a package
// policy body that is common to every resource that creates or updates one
// (elasticstack_fleet_integration_policy today; elasticstack_fleet_managed_integration).
//
// # Naming
//
// This package started life as the "policyshape" working name proposed in
// openspec/changes/fleet-agentless-policy/design.md (Decision 2) and
// openspec/changes/fleet-agentless-policy/specs/fleet-policyshape/spec.md.
// The name was validated against the existing internal/fleet/* naming
// convention before being finalized:
//
//   - Newer Fleet sub-packages are named after the concept they model, as a
//     single lowercase word with no separators: agentpolicy, integrationds,
//     outputds, customintegration, serverhost, enrollmenttokens, proxy.
//   - Older sub-packages (integration_policy, elastic_defend_integration_policy)
//     use snake_case and predate that convention.
//
// "policyshape" fits the newer, terser convention: it names the concept
// (the shared shape of a package-policy body — inputs/streams/vars) rather
// than any one resource that consumes it, which matters because this
// package is explicitly designed to be imported by more than one resource
// package. No better alternative (e.g. "packagepolicy", "policymodel") was
// found to be meaningfully clearer, so the working name was kept as final.
//
// # What lives here vs. what doesn't
//
// This package owns:
//   - InputType/InputValue and InputsType/InputsValue: the Plugin Framework
//     custom types for a single input (and its nested streams) and for the
//     top-level inputs map, including defaults-aware semantic equality.
//   - VarsJSONType/VarsJSONValue: the custom type for the top-level
//     (integration-level) vars_json attribute, with package-default
//     population via a caller-supplied lookup function.
//   - Defaults merging: deriving default var/stream values from Fleet
//     package metadata (policy templates, data streams) and merging them
//     with user-supplied values.
//   - Secret helpers: detecting secret-reference vars and preserving them
//     across updates without ever surfacing the raw value in state.
//
// This package deliberately does NOT own:
//   - Generic JSON string normalization — that already lives in
//     internal/utils/typeutils (MarshalToNormalized, NormalizeJSONScalar) and
//     internal/utils/customtypes (JSONWithContextualDefaultsType), which are
//     provider-wide utilities, not Fleet-specific. VarsJSONType is a thin,
//     Fleet-shaped wrapper around that existing machinery.
//   - Terraform resource.Schema construction (descriptions, sensitivity
//     flags, validators) — each resource owns its own schema.go and decides
//     how to present these types to users.
//   - Package-info caching (the sync.Map keyed by "<name>-<version>") — each
//     resource owns its own cache and passes a lookup callback into this
//     package where needed, so resources don't share unrelated cache state.
package policyshape
