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

// Bridges the dashboard-level populatePanelConfigJSONDefaults into this package without importing the
// dashboard package (which would introduce a circular dependency).
// Section 5 of dashboard-panel-contract routes JSON defaulting through Handler.PopulateJSONDefaults and
// deletes this bridging approach.
package sloerrorbudget

import (
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
)

var panelJSONPopulateDefaults customtypes.PopulateDefaultsFunc[map[string]any]

// SetPopulatePanelConfigJSONDefaults binds the dashboard package panel config_json default-population
// function so panel ConfigJSON values remain semantically comparable to other dashboard code paths.
// Must be invoked from dashboard package init before handlers run.
func SetPopulatePanelConfigJSONDefaults(fn customtypes.PopulateDefaultsFunc[map[string]any]) {
	panelJSONPopulateDefaults = fn
}

func panelConfigJSONNull() customtypes.JSONWithDefaultsValue[map[string]any] {
	fn := panelJSONPopulateDefaults
	if fn == nil {
		fn = func(m map[string]any) map[string]any { return m }
	}
	return customtypes.NewJSONWithDefaultsNull(fn)
}
