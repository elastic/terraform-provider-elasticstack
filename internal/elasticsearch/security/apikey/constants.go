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

package apikey

// Package-level constants shared between the resource and ephemeral resource.
const (
	// CurrentSchemaVersion is the resource schema version. It is bumped when a
	// state upgrader is added to handle a breaking schema change.
	CurrentSchemaVersion int64 = 2

	// RESTAPIKeyType is the value the `type` attribute takes for standard API
	// keys.
	RESTAPIKeyType = "rest"

	// CrossClusterAPIKeyType is the value the `type` attribute takes for
	// cross-cluster API keys.
	CrossClusterAPIKeyType = "cross_cluster"

	// DefaultAPIKeyType is applied when the `type` attribute is unset.
	DefaultAPIKeyType = RESTAPIKeyType

	// APIKeyNameInvalidMessage is the validator error message used when the
	// `name` attribute violates the allowed-character constraints.
	APIKeyNameInvalidMessage = "must contain alphanumeric characters (a-z, A-Z, 0-9), spaces, punctuation, and printable symbols in the Basic Latin (ASCII) block. " +
		"Leading or trailing whitespace is not allowed"
)
