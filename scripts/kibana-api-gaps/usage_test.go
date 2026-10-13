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

package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type usageProbe struct {
	t     *testing.T
	usage *usageSet
	mod   *loadedModule
}

func usageFor(t *testing.T, pkg string) usageProbe {
	t.Helper()
	mod := loadFixture(t, "testdata/fixture")
	return usageProbe{t: t, mod: mod, usage: collectUsage(computeReach(mod, fixtureKbapi, "fixture/"+pkg))}
}

func (p usageProbe) fieldUsed(typeName, field string) bool {
	p.t.Helper()
	obj := p.mod.pkg(p.t, fixtureKbapi).Types.Scope().Lookup(typeName)
	require.NotNil(p.t, obj, typeName)
	entry, ok := jsonFields(obj.Type())[field]
	require.True(p.t, ok, field)
	return p.usage.fieldUsed(entry.owner, entry.field.Name())
}

func (p usageProbe) opaque(typeName string) bool {
	p.t.Helper()
	obj := p.mod.pkg(p.t, fixtureKbapi).Types.Scope().Lookup(typeName)
	require.NotNil(p.t, obj, typeName)
	return p.usage.opaque(obj.Type())
}

func TestUsageCountsCompositeLiteralKeysAndAssignedFields(t *testing.T) {
	p := usageFor(t, "direct")
	assert.True(t, p.fieldUsed("Widget", "name"))
	assert.True(t, p.fieldUsed("Widget", "labels"))
	assert.True(t, p.fieldUsed("Widget", "settings"))
	assert.True(t, p.fieldUsed("Label", "key"))
	assert.True(t, p.fieldUsed("Settings", "mode"))
	assert.True(t, p.fieldUsed("CreateWidgetParams", "dry_run"))
}

func TestUsageLeavesUntouchedFieldsUnused(t *testing.T) {
	p := usageFor(t, "direct")
	assert.False(t, p.fieldUsed("Widget", "description"))
	assert.False(t, p.fieldUsed("Label", "value"))
	assert.False(t, p.fieldUsed("Settings", "retries"))
	assert.False(t, p.fieldUsed("CreateWidgetParams", "verbose"))
}

func TestUsageCountsFieldsUsedInCallbacksAndSelectorAssignments(t *testing.T) {
	p := usageFor(t, "callback")
	assert.True(t, p.fieldUsed("Widget", "description"))
	assert.True(t, p.fieldUsed("Widget", "settings"))
	assert.True(t, p.fieldUsed("Settings", "retries"))
	assert.False(t, p.fieldUsed("Settings", "mode"))
}

func TestUsageCountsFieldsUsedInSharedHelpers(t *testing.T) {
	p := usageFor(t, "helperuser")
	assert.True(t, p.fieldUsed("Widget", "description"))
	assert.False(t, p.fieldUsed("Widget", "labels"))
}

func TestUsageTreatsWholeValueJSONPassthroughAsOpaque(t *testing.T) {
	p := usageFor(t, "passthrough")
	assert.True(t, p.opaque("Widget"))
	assert.False(t, usageFor(t, "direct").opaque("Widget"))
}

func TestUsageTreatsTypesPassedToReflectThroughGenericHelpersAsOpaque(t *testing.T) {
	assert.True(t, usageFor(t, "reflecting").opaque("Widget"))
}

func TestUsageTreatsTypesPassedToReflectThroughInterfaceParametersAsOpaque(t *testing.T) {
	p := usageFor(t, "reflectany")
	assert.True(t, p.opaque("Widget"))
	assert.True(t, p.opaque("Settings"), "a concrete value handed straight to reflect is opaque")
	assert.False(t, usageFor(t, "direct").opaque("Settings"))
}

func TestUsageFollowsReflectionThroughWrappersAndInterfaceCalls(t *testing.T) {
	p := usageFor(t, "reflectwrap")
	assert.True(t, p.opaque("Widget"), "wrapper forwarding to a reflecting helper")
	assert.True(t, p.opaque("Settings"), "interface call resolved to a reflecting implementation")
}
