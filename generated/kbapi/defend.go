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

// This file is hand-authored, not generated. It extends the kbapi package
// with Elastic Defend-specific types that use the typed-input encoding.
// It will NOT be overwritten by make generate, but must be maintained
// manually if the Fleet package policy API changes.
//
// Background: the generated kibana.gen.go uses a map-keyed input shape
// (simplified format) which does not match the Defend API. These types provide
// the typed-input shape required for Defend bootstrap and finalize operations.

package kbapi

// DefendPackagePolicyRequestInput is a typed input entry used by the Elastic
// Defend package policy. The map key (input type, e.g. "endpoint") acts as the
// type discriminator; the Fleet API expects inputs as a JSON object (map) rather
// than an array.
type DefendPackagePolicyRequestInput struct {
	// Enabled indicates whether the input is active.
	Enabled bool `json:"enabled"`

	// Streams is fixed to an empty object for Defend inputs.
	// The Fleet simplified-format schema expects streams as a map (object), not an array.
	Streams map[string]any `json:"streams"`

	// Config holds the Defend-specific input configuration.
	// Keys include "integration_config" (with "value" wrapper), "artifact_manifest",
	// and "policy" (with "value" wrapper).
	Config map[string]any `json:"config,omitempty"`
}

// DefendPackagePolicyRequest is the request body for creating or updating an
// Elastic Defend package policy. It uses the same map-keyed input format as the
// generic PackagePolicyRequest (format=simplified), where the map key is the
// input type (e.g. "endpoint").
type DefendPackagePolicyRequest struct {
	// Id is the package policy unique identifier (optional on create, required on update).
	Id *string `json:"id,omitempty"`

	// Name is the unique name for the policy.
	Name string `json:"name"`

	// Namespace is the policy namespace.
	Namespace *string `json:"namespace,omitempty"`

	// Description is the policy description.
	Description *string `json:"description,omitempty"`

	// Enabled indicates whether the policy is enabled.
	Enabled *bool `json:"enabled,omitempty"`

	// Force forces creation/deletion even if the package is not verified or
	// the agent policy is managed.
	Force *bool `json:"force,omitempty"`

	// Package identifies the integration package.
	Package PackagePolicyRequestPackage `json:"package"`

	// PolicyId is the agent policy ID (deprecated; prefer PolicyIds).
	PolicyId *string `json:"policy_id,omitempty"`

	// Version is the Elasticsearch version token required for Defend update
	// requests. It is populated from the package policy response and must be
	// echoed back unchanged.
	Version *string `json:"version,omitempty"`

	// Inputs is the map-keyed set of Defend inputs, keyed by input type (e.g. "endpoint").
	// The Fleet API expects inputs as a JSON object, not an array.
	Inputs map[string]DefendPackagePolicyRequestInput `json:"inputs"`
}

// DefendPackagePolicyInput is a typed input entry in the Elastic Defend package
// policy response. Unlike the generic mapped input, each Defend input carries an
// explicit "type" discriminator and "config" payload.
type DefendPackagePolicyInput struct {
	// Type is the input type discriminator ("endpoint").
	Type string `json:"type"`

	// Enabled indicates whether the input is active.
	Enabled bool `json:"enabled"`

	// Config holds the Defend-specific input configuration, including
	// "integration_config", "artifact_manifest", "policy", and other fields.
	Config map[string]any `json:"config,omitempty"`
}

// DefendPackagePolicy is the response body for an Elastic Defend package policy.
// It uses a typed input list rather than the simplified mapped-input format.
type DefendPackagePolicy struct {
	// Id is the package policy unique identifier.
	Id string `json:"id"`

	// Name is the unique name for the package policy.
	Name string `json:"name"`

	// Namespace is the policy namespace.
	Namespace *string `json:"namespace,omitempty"`

	// Description is the policy description.
	Description *string `json:"description,omitempty"`

	// Enabled indicates whether the policy is enabled.
	Enabled bool `json:"enabled"`

	// Package identifies the integration package.
	Package *struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	} `json:"package,omitempty"`

	// PolicyId is the agent policy ID.
	PolicyId *string `json:"policy_id,omitempty"`

	// PolicyIds is the list of agent policy IDs.
	PolicyIds *[]string `json:"policy_ids,omitempty"`

	// SpaceIds is the list of Kibana space IDs the policy belongs to.
	SpaceIds *[]string `json:"spaceIds,omitempty"`

	// Version is the Elasticsearch version token required for updates.
	Version *string `json:"version,omitempty"`

	// Inputs is the typed list of Defend inputs.
	Inputs []DefendPackagePolicyInput `json:"inputs"`
}
