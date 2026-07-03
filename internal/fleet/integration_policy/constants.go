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

// Terraform schema attribute keys shared across integration policy schema
// versions (V0–V3) and nested inputs/streams attribute type maps.
const (
	attrPolicyID           = "policy_id"
	attrName               = "name"
	attrNamespace          = "namespace"
	attrAgentPolicyID      = "agent_policy_id"
	attrAgentPolicyIDs     = "agent_policy_ids"
	attrDescription        = "description"
	attrForce              = "force"
	attrIntegrationName    = "integration_name"
	attrIntegrationVersion = "integration_version"
	attrOutputID           = "output_id"
	attrVarsJSON           = "vars_json"
	attrSpaceIDs           = "space_ids"
	attrEnabled            = "enabled"
	attrVars               = "vars"
	attrDefaults           = "defaults"
	attrStreams            = "streams"
	attrCondition          = "condition"
)
