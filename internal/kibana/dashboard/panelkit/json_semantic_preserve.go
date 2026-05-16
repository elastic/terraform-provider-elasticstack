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
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// PreservePriorJSONWithDefaultsIfEquivalent returns prior when it is semantically equal to current
// (via JSON-with-defaults comparison), otherwise current. Used on dashboard panel reads to avoid
// state churn when Kibana echoes defaults.
func PreservePriorJSONWithDefaultsIfEquivalent[T any](ctx context.Context, prior, current customtypes.JSONWithDefaultsValue[T], diags *diag.Diagnostics) customtypes.JSONWithDefaultsValue[T] {
	if prior.IsNull() || prior.IsUnknown() || current.IsNull() || current.IsUnknown() {
		return current
	}

	eq, d := prior.StringSemanticEquals(ctx, current)
	diags.Append(d...)
	if d.HasError() {
		return current
	}
	if eq {
		return prior
	}
	return current
}

// PreservePriorNormalizedWithDefaultsIfEquivalent returns prior when it is semantically equal to
// current after applying defaults to both sides (JSON-with-defaults comparison over jsontypes.Normalized
// payloads). Used on dashboard panel reads to avoid state churn when Kibana echoes defaults inside
// fields the panel handler stores as plain normalized JSON.
func PreservePriorNormalizedWithDefaultsIfEquivalent[T any](ctx context.Context, prior, current jsontypes.Normalized, defaults func(T) T, diags *diag.Diagnostics) jsontypes.Normalized {
	if prior.IsNull() || prior.IsUnknown() || current.IsNull() || current.IsUnknown() {
		return current
	}

	priorWithDefaults := customtypes.NewJSONWithDefaultsValue(prior.ValueString(), defaults)
	currentWithDefaults := customtypes.NewJSONWithDefaultsValue(current.ValueString(), defaults)
	eq, d := priorWithDefaults.StringSemanticEquals(ctx, currentWithDefaults)
	diags.Append(d...)
	if d.HasError() {
		return current
	}
	if eq {
		return prior
	}
	return current
}
