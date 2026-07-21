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

import "github.com/hashicorp/terraform-plugin-framework/diag"

// ApplySimpleConfig handles the three-guard null-preservation preamble shared by every AIOps-style
// PopulateFromAPI function:
//
//  1. Import path (priorCfg == nil or *priorCfg == nil): populate *dst from the API via factory and return.
//  2. Type-change recovery (*dst == nil but *priorCfg != nil): same — rebuild from API and return.
//  3. Existing-nil guard (*dst == nil after prior check): nothing to do, return immediately.
//
// After the guards pass, populateFn receives the non-nil existing *C and the api value so the caller
// can fill in panel-specific fields. populateFn may return diagnostics; they are forwarded to the
// caller.
//
// dst and priorCfg are pointers to the config fields on the current and prior PanelModel structs
// (e.g. &pm.AiopsChangePointChartConfig and &prior.AiopsChangePointChartConfig). priorCfg may be
// nil when prior is nil (import path).
func ApplySimpleConfig[C any, A any](
	dst **C,
	priorCfg **C,
	api A,
	factory func(A) *C,
	populateFn func(*C, A) diag.Diagnostics,
) diag.Diagnostics {
	// Import path: no prior state at all.
	if priorCfg == nil || *priorCfg == nil {
		*dst = factory(api)
		return nil
	}

	// Type-change recovery: config block disappeared from the plan but was present in prior state.
	if *dst == nil {
		*dst = factory(api)
		return nil
	}

	// Existing-nil guard: config is intentionally absent in the current plan.
	existing := *dst
	if existing == nil {
		return nil
	}

	return populateFn(existing, api)
}
