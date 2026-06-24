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

// Package agentbuilder provides shared constants for the Agent Builder Kibana
// resources. Version gates are centralised here so that all sub-packages
// reference a single, grep-friendly source of truth.
//
// API rollout timeline:
//   - 9.3.0 GA  – core resources: agents and tools
//   - 9.4.0     – extended resources: skills, workflows, and advanced agent
//     config (workflow_ids / skill_ids / plugin_ids on agents)
package agentbuilder

import semver "github.com/hashicorp/go-version"

var (
	// MinCoreAPIVersion is the minimum Kibana version required for the Agent
	// Builder core resources: agents and tools.
	MinCoreAPIVersion = semver.Must(semver.NewVersion("9.3.0"))

	// MinExtendedAPIVersion is the minimum Kibana version required for the
	// Agent Builder extended resources: skills, workflows, and advanced agent
	// config (workflow_ids, skill_ids, plugin_ids).
	MinExtendedAPIVersion = semver.Must(semver.NewVersion("9.4.0-SNAPSHOT"))
)
