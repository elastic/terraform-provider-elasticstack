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

package index

// Terraform schema attribute keys reused by the index resource and its sort
// support. These names match the Elasticsearch index sort and alias APIs.
const (
	attrName    = "name"
	attrFilter  = "filter"
	attrSetting = "setting"
	attrField   = "field"
	attrOrder   = "order"
	attrMissing = "missing"
	attrMode    = "mode"
)

// Sort order and tie-breaker tokens used in index sort plan modifiers and
// flattened state.
// importHydrationPrivateStateKey is set during ImportState and consumed on the
// following Read (hydrate all settings) and ModifyPlan (prune unconfigured fields).
const importHydrationPrivateStateKey = "import_hydration"

const (
	sortOrderAsc    = "asc"
	sortOrderDesc   = "desc"
	sortMissingLast = "_last"
	sortModeMax     = "max"
	sortModeMin     = "min"
)
