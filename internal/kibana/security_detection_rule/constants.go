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

package securitydetectionrule

// Rule type discriminators recognised by the Kibana Security Detections API.
const (
	ruleTypeQuery           = "query"
	ruleTypeEQL             = "eql"
	ruleTypeESQL            = "esql"
	ruleTypeMachineLearning = "machine_learning"
	ruleTypeNewTerms        = "new_terms"
	ruleTypeSavedQuery      = "saved_query"
	ruleTypeThreatMatch     = "threat_match"
	ruleTypeThreshold       = "threshold"
)

// Terraform schema attribute keys that appear in multiple nested blocks of the
// security detection rule schema. They are pulled out as constants to satisfy
// goconst and provide a single point of truth for shared attribute names.
const (
	attrName         = "name"
	attrType         = "type"
	attrQuery        = "query"
	attrField        = "field"
	attrValue        = "value"
	attrVersion      = "version"
	attrActions      = "actions"
	attrActionTypeID = "action_type_id"
	attrParams       = "params"
	attrKQL          = "kql"
	attrReference    = "reference"
)
