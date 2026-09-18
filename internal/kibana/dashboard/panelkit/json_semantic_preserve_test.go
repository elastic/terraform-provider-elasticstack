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

package panelkit_test

import (
	"context"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/stretchr/testify/assert"
)

func identityDefaults(m map[string]any) map[string]any {
	return m
}

func TestPreservePriorJSONWithDefaultsIfEquivalent(t *testing.T) {
	t.Parallel()

	t.Run("semantically equal prior and current -> current reverts to prior's exact formatting", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		// Prior has extra whitespace but is semantically identical to current once defaults are
		// applied. An exact-string check (not JSONEq) proves prior was preserved verbatim rather
		// than current being left as-is.
		priorJSON := `{"field": "a"}`
		prior := customtypes.NewJSONWithDefaultsValue(priorJSON, identityDefaults)
		current := customtypes.NewJSONWithDefaultsValue(`{"field":"a"}`, identityDefaults)
		var diags diag.Diagnostics

		result := panelkit.PreservePriorJSONWithDefaultsIfEquivalent(ctx, prior, current, &diags)

		assert.Equal(t, priorJSON, result.ValueString()) //nolint:testifylint
	})

	t.Run("semantically different prior and current -> current is left unchanged", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		prior := customtypes.NewJSONWithDefaultsValue(`{"field":"a"}`, identityDefaults)
		currentJSON := `{"field":"different"}`
		current := customtypes.NewJSONWithDefaultsValue(currentJSON, identityDefaults)
		var diags diag.Diagnostics

		result := panelkit.PreservePriorJSONWithDefaultsIfEquivalent(ctx, prior, current, &diags)

		assert.Equal(t, currentJSON, result.ValueString())
	})

	t.Run("unknown prior -> current is left unchanged", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		prior := customtypes.NewJSONWithDefaultsUnknown[map[string]any](identityDefaults)
		currentJSON := `{"field":"a"}`
		current := customtypes.NewJSONWithDefaultsValue(currentJSON, identityDefaults)
		var diags diag.Diagnostics

		result := panelkit.PreservePriorJSONWithDefaultsIfEquivalent(ctx, prior, current, &diags)

		assert.Equal(t, currentJSON, result.ValueString())
	})

	t.Run("null current -> current is left unchanged", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		prior := customtypes.NewJSONWithDefaultsValue(`{"field":"a"}`, identityDefaults)
		current := customtypes.NewJSONWithDefaultsNull[map[string]any](identityDefaults)
		var diags diag.Diagnostics

		result := panelkit.PreservePriorJSONWithDefaultsIfEquivalent(ctx, prior, current, &diags)

		assert.True(t, result.IsNull())
	})
}
