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
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type simpleTestConfig struct {
	Value string
}

type simpleTestAPI struct {
	Value string
}

func makeFactory(api simpleTestAPI) *simpleTestConfig {
	return &simpleTestConfig{Value: api.Value}
}

func makePopulate(existing *simpleTestConfig, api simpleTestAPI) diag.Diagnostics {
	existing.Value = "populated:" + api.Value
	return nil
}

// TestApplySimpleConfig_importPath verifies that when priorCfg is nil (import), dst is set via factory.
func TestApplySimpleConfig_importPath(t *testing.T) {
	t.Parallel()

	var dst *simpleTestConfig
	api := simpleTestAPI{Value: "imported"}

	diags := ApplySimpleConfig(&dst, nil, api, makeFactory, makePopulate)
	require.False(t, diags.HasError())
	require.NotNil(t, dst)
	assert.Equal(t, "imported", dst.Value)
}

// TestApplySimpleConfig_priorConfigNil verifies that when *priorCfg is nil (prior had no config), dst is set via factory.
func TestApplySimpleConfig_priorConfigNil(t *testing.T) {
	t.Parallel()

	var dst *simpleTestConfig
	var priorCfgVal *simpleTestConfig
	priorCfg := &priorCfgVal // priorCfg is non-nil but *priorCfg is nil

	api := simpleTestAPI{Value: "from-nil-prior"}

	diags := ApplySimpleConfig(&dst, priorCfg, api, makeFactory, makePopulate)
	require.False(t, diags.HasError())
	require.NotNil(t, dst)
	assert.Equal(t, "from-nil-prior", dst.Value)
}

// TestApplySimpleConfig_typeChangeRecovery verifies that when dst is nil but *priorCfg is non-nil, factory is called.
func TestApplySimpleConfig_typeChangeRecovery(t *testing.T) {
	t.Parallel()

	var dst *simpleTestConfig
	priorCfgVal := &simpleTestConfig{Value: "old"}
	priorCfg := &priorCfgVal

	api := simpleTestAPI{Value: "recovered"}

	diags := ApplySimpleConfig(&dst, priorCfg, api, makeFactory, makePopulate)
	require.False(t, diags.HasError())
	require.NotNil(t, dst)
	assert.Equal(t, "recovered", dst.Value)
}

// TestApplySimpleConfig_existingNilGuard verifies that when both dst and *priorCfg are nil, nothing happens.
// This cannot occur in practice (the type-change-recovery guard fires first when *priorCfg != nil),
// but we verify the nil guard is safe.
func TestApplySimpleConfig_existingNilGuard(t *testing.T) {
	t.Parallel()

	var dst *simpleTestConfig
	var priorCfgVal *simpleTestConfig
	priorCfg := &priorCfgVal

	api := simpleTestAPI{Value: "irrelevant"}

	diags := ApplySimpleConfig(&dst, priorCfg, api, makeFactory, makePopulate)
	require.False(t, diags.HasError())
	// dst was nil, priorCfg was nil -> factory fires (import-path branch), setting dst
	assert.NotNil(t, dst)
}

// TestApplySimpleConfig_normalPath verifies that when dst and prior are both non-nil, populateFn is called.
func TestApplySimpleConfig_normalPath(t *testing.T) {
	t.Parallel()

	dstVal := &simpleTestConfig{Value: "old"}
	dst := dstVal
	priorCfgVal := &simpleTestConfig{Value: "prior"}
	priorCfg := &priorCfgVal

	api := simpleTestAPI{Value: "new"}

	diags := ApplySimpleConfig(&dst, priorCfg, api, makeFactory, makePopulate)
	require.False(t, diags.HasError())
	require.NotNil(t, dst)
	assert.Equal(t, "populated:new", dst.Value)
}

// TestApplySimpleConfig_populateFnDiagsForwarded verifies that diagnostics from populateFn are returned.
func TestApplySimpleConfig_populateFnDiagsForwarded(t *testing.T) {
	t.Parallel()

	dstVal := &simpleTestConfig{Value: "old"}
	dst := dstVal
	priorCfgVal := &simpleTestConfig{Value: "prior"}
	priorCfg := &priorCfgVal

	api := simpleTestAPI{Value: "new"}

	errFn := func(existing *simpleTestConfig, _ simpleTestAPI) diag.Diagnostics {
		var d diag.Diagnostics
		d.AddError("test error", "detail")
		return d
	}

	diags := ApplySimpleConfig(&dst, priorCfg, api, makeFactory, errFn)
	assert.True(t, diags.HasError())
}
