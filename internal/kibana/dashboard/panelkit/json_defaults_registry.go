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

package panelkit

import (
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
)

// GlobalPanelJSONDefaults is set once by dashboard init and shared by all panel packages.
// Panel packages must not import the dashboard package (circular dependency), so this
// shared variable in panelkit acts as the bridge.
var GlobalPanelJSONDefaults customtypes.PopulateDefaultsFunc[map[string]any]

// PanelJSONDefaultsFunc returns GlobalPanelJSONDefaults, falling back to an identity
// function when it has not been set (e.g. in unit tests).
func PanelJSONDefaultsFunc() customtypes.PopulateDefaultsFunc[map[string]any] {
	fn := GlobalPanelJSONDefaults
	if fn == nil {
		fn = func(m map[string]any) map[string]any { return m }
	}
	return fn
}

// PanelConfigJSONNull returns a null JSONWithDefaultsValue wired with PanelJSONDefaultsFunc.
func PanelConfigJSONNull() customtypes.JSONWithDefaultsValue[map[string]any] {
	return customtypes.NewJSONWithDefaultsNull(PanelJSONDefaultsFunc())
}
