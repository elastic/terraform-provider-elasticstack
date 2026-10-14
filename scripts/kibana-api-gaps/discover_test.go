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
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	widgetPtr   = "/components/schemas/Widget/properties/"
	settingsPtr = "/components/schemas/Settings/properties/"
)

var fixtureResources = []Resource{
	{Name: "fx_direct", PkgPath: "fixture/direct"},
	{Name: "fx_callback", PkgPath: "fixture/callback"},
	{Name: "fx_helperuser", PkgPath: "fixture/helperuser"},
	{Name: "fx_passthrough", PkgPath: "fixture/passthrough"},
	{Name: "fx_nokibana", PkgPath: "fixture/nokibana"},
	{Name: "fx_blobs", PkgPath: "fixture/blobs"},
	{Name: "fx_rawbody", PkgPath: "fixture/rawbody"},
	{Name: "fx_nohide", PkgPath: "fixture/nohide"},
	{Name: "fx_convert", PkgPath: "fixture/convert"},
	{Name: "fx_copyassign", PkgPath: "fixture/copyassign"},
	{Name: "fx_unionpass", PkgPath: "fixture/unionpass"},
	{Name: "fx_reflecting", PkgPath: "fixture/reflecting"},
	{Name: "fx_reflectany", PkgPath: "fixture/reflectany"},
}

func fixtureOASVariant(t *testing.T, oldText, newText string) string {
	t.Helper()
	data, err := os.ReadFile(fixtureOAS)
	require.NoError(t, err)
	require.Contains(t, string(data), oldText)
	path := filepath.Join(t.TempDir(), "oas.yaml")
	require.NoError(t, os.WriteFile(path, []byte(strings.Replace(string(data), oldText, newText, 1)), 0o600))
	return path
}

func discoverFixture(t *testing.T, oasPath string) (*Discovery, error) {
	t.Helper()
	doc, err := loadOAS(oasPath)
	require.NoError(t, err)
	return discover(discoverConfig{module: loadFixture(t, "testdata/fixture"), kbapiPath: fixtureKbapi, oas: doc, resources: fixtureResources})
}

func mustDiscover(t *testing.T) *Discovery {
	t.Helper()
	d, err := discoverFixture(t, fixtureOAS)
	require.NoError(t, err)
	return d
}

func (d *Discovery) group(key string) *Group {
	for i := range d.Groups {
		if d.Groups[i].Key == key {
			return &d.Groups[i]
		}
	}
	return nil
}

func memberFields(g *Group) []string {
	var fields []string
	for _, m := range g.Members {
		fields = append(fields, m.Field)
	}
	return fields
}

func TestDiscoverReportsUnusedRequestFieldAsCandidate(t *testing.T) {
	g := mustDiscover(t).group("fx_direct#" + widgetPtr + "description")
	require.NotNil(t, g)
	assert.Equal(t, []string{"description"}, memberFields(g))
	assert.Equal(t, "Generally available in 8.15.0", g.Members[0].XState)
}

func TestDiscoverDoesNotReportUsedFields(t *testing.T) {
	d := mustDiscover(t)
	for _, name := range []string{"name", "labels.key"} {
		for _, g := range d.Groups {
			if g.Resource == "fx_direct" {
				assert.NotContains(t, memberFields(&g), name)
			}
		}
	}
	assert.Nil(t, d.group("fx_direct#"+widgetPtr+"name"))
	assert.Nil(t, d.group("fx_direct#/paths/~1api~1widgets/post/parameters/dry_run"), "dry_run is used")
}

func TestDiscoverReportsUnusedQueryParametersAsTopLevelFields(t *testing.T) {
	g := mustDiscover(t).group("fx_direct#/paths/~1api~1widgets/post/parameters/verbose")
	require.NotNil(t, g)
	assert.Equal(t, []string{"verbose"}, memberFields(g))
}

func TestDiscoverExcludesDeprecatedAndResponseOnlyFields(t *testing.T) {
	d := mustDiscover(t)
	assert.Nil(t, d.group("fx_direct#"+widgetPtr+"legacy"))
	assert.Nil(t, d.group("fx_direct#"+widgetPtr+"id"))
}

func TestDiscoverGroupsNestedUnusedFieldsUnderTheirTopLevelField(t *testing.T) {
	d := mustDiscover(t)

	settings := d.group("fx_direct#" + widgetPtr + "settings")
	require.NotNil(t, settings)
	assert.Equal(t, []string{"settings.retries"}, memberFields(settings))
	assert.Equal(t, settingsPtr+"retries", settings.Members[0].Pointer)

	labels := d.group("fx_direct#" + widgetPtr + "labels")
	require.NotNil(t, labels)
	assert.Equal(t, []string{"labels[].value"}, memberFields(labels))
}

func TestDiscoverReportsWholeUnusedTopLevelFieldAsOneMember(t *testing.T) {
	g := mustDiscover(t).group("fx_direct#" + widgetPtr + "source")
	require.NotNil(t, g)
	assert.Equal(t, []string{"source"}, memberFields(g))
}

func TestDiscoverWalksOnlyUsedUnionBranches(t *testing.T) {
	g := mustDiscover(t).group("fx_callback#" + widgetPtr + "source")
	require.NotNil(t, g)
	assert.Equal(t, []string{"source.b"}, memberFields(g))
}

func TestDiscoverCountsFieldsUsedByCallbacksAndSharedHelpers(t *testing.T) {
	d := mustDiscover(t)
	assert.Nil(t, d.group("fx_callback#"+widgetPtr+"description"))
	assert.Nil(t, d.group("fx_helperuser#"+widgetPtr+"description"))
	assert.NotNil(t, d.group("fx_helperuser#"+widgetPtr+"labels"))
}

func TestDiscoverTreatsOpaquePassthroughAsCovered(t *testing.T) {
	for _, g := range mustDiscover(t).Groups {
		if g.Resource == "fx_passthrough" {
			assert.NotContains(t, g.Pointer, "/components/schemas/Widget/", g.Key)
		}
	}
}

func TestDiscoverSkipsResourcesThatDoNotCallTheKibanaClient(t *testing.T) {
	d := mustDiscover(t)
	for _, g := range d.Groups {
		assert.NotEqual(t, "fx_nokibana", g.Resource)
	}
	assert.Contains(t, d.NoOperations, "fx_nokibana")
	assert.Equal(t, 1, d.Operations["fx_direct"])
	assert.Equal(t, 1, d.Operations["fx_helperuser"])
}

func TestDiscoverIgnoresEndpointsNoResourceCalls(t *testing.T) {
	for _, g := range mustDiscover(t).Groups {
		assert.NotContains(t, g.Pointer, "gadgets", g.Key)
	}
}

func TestDiscoverIsRepeatable(t *testing.T) {
	first, err := json.Marshal(mustDiscover(t))
	require.NoError(t, err)
	second, err := json.Marshal(mustDiscover(t))
	require.NoError(t, err)
	assert.Equal(t, first, second)
}

func TestDiscoverFailsOnJSONTagWithoutMatchingGoField(t *testing.T) {
	oas := fixtureOASVariant(t, "        name: {type: string}\n", "        name: {type: string}\n        ghost: {type: string}\n")
	_, err := discoverFixture(t, oas)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ghost")
}

func TestDiscoverRecordsUnsupportedShapesAndTreatsThemAsCovered(t *testing.T) {
	oas := fixtureOASVariant(t, "        description: {type: string, x-state: Generally available in 8.15.0}\n",
		"        description: {not: {type: integer}}\n")
	d, err := discoverFixture(t, oas)
	require.NoError(t, err)
	assert.Nil(t, d.group("fx_direct#"+widgetPtr+"description"))
	require.NotEmpty(t, d.SkippedShapes)
	assert.Equal(t, widgetPtr+"description", d.SkippedShapes[0].Pointer)
	assert.Contains(t, d.SkippedShapes[0].Shape, "not")
}

func TestDiscoverGroupFingerprintFollowsStructureNotDescriptions(t *testing.T) {
	key := "fx_direct#" + widgetPtr + "settings"
	base := mustDiscover(t).group(key).Fingerprint

	described, err := discoverFixture(t, fixtureOASVariant(t, "        retries: {type: integer}\n", "        retries: {type: integer, description: how often}\n"))
	require.NoError(t, err)
	assert.Equal(t, base, described.group(key).Fingerprint)

	restructured, err := discoverFixture(t, fixtureOASVariant(t, "mode: {type: string, enum: [fast, slow]}", "mode: {type: string, enum: [fast, slow, turbo]}"))
	require.NoError(t, err)
	assert.NotEqual(t, base, restructured.group(key).Fingerprint)
}

func TestDiscoverWalksUnionRequestBodiesBranchByBranch(t *testing.T) {
	d := mustDiscover(t)
	text := d.group("fx_blobs#/components/schemas/BlobText/properties/text")
	require.NotNil(t, text)
	assert.Equal(t, []string{"text"}, memberFields(text))
	assert.NotNil(t, d.group("fx_blobs#/components/schemas/BlobNumber/properties/number"))
}

func TestDiscoverWalksRawBodyClientMethodsWhenTheTypedBodyIsKnown(t *testing.T) {
	d := mustDiscover(t)
	assert.NotNil(t, d.group("fx_rawbody#"+widgetPtr+"description"))
	assert.NotNil(t, d.group("fx_rawbody#/paths/~1api~1widgets/post/parameters/dry_run"))
	assert.Nil(t, d.group("fx_rawbody#/paths/~1api~1widgets/post/parameters/verbose"))
}

func TestDiscoverSkipsRawBodiesWithoutATypedBody(t *testing.T) {
	doc, err := loadOAS(fixtureOAS)
	require.NoError(t, err)
	op, ok := doc.operation("POST", "/api/blobs")
	require.True(t, ok)
	w := newWalker(doc, &usageSet{}, nil, "r")

	_, err = operationGaps(doc, w, opSpec{Method: "POST", Path: "/api/blobs", RawBody: true})
	require.NoError(t, err)
	require.Len(t, w.skipped, 1)
	assert.Equal(t, op.Body.pointer, w.skipped[0].Pointer)
}

func TestDiscoverDoesNotHideGapsBehindLocalCopiesOrLiteralConversions(t *testing.T) {
	d := mustDiscover(t)
	labels := d.group("fx_nohide#" + widgetPtr + "labels")
	require.NotNil(t, labels)
	assert.Equal(t, []string{"labels[].value"}, memberFields(labels))
	settings := d.group("fx_nohide#" + widgetPtr + "settings")
	require.NotNil(t, settings)
	assert.Equal(t, []string{"settings.retries"}, memberFields(settings))
}

func TestDiscoverKeysQueryParameterGroupsByNameNotPosition(t *testing.T) {
	reordered := fixtureOASVariant(t,
		"        - {name: dry_run, in: query, schema: {type: boolean}}\n        - {name: verbose, in: query, schema: {type: boolean}}\n",
		"        - {name: verbose, in: query, schema: {type: boolean}}\n        - {name: dry_run, in: query, schema: {type: boolean}}\n")
	d, err := discoverFixture(t, reordered)
	require.NoError(t, err)
	assert.NotNil(t, d.group("fx_direct#/paths/~1api~1widgets/post/parameters/verbose"))
}

func TestDiscoverTreatsWholeValueConversionsCopiesAndUnionPassthroughAsCovered(t *testing.T) {
	d := mustDiscover(t)
	for _, resource := range []string{"fx_convert", "fx_copyassign", "fx_unionpass"} {
		for _, g := range d.Groups {
			if g.Resource == resource {
				assert.NotContains(t, g.Pointer, "/components/schemas/", g.Key)
			}
		}
	}
}

func TestAddGroupMergesRequiredAcrossOperationsRegardlessOfOrder(t *testing.T) {
	doc, err := loadOAS(fixtureOAS)
	require.NoError(t, err)
	node, err := doc.lookup(widgetPtr + "name")
	require.NoError(t, err)
	prop := func(required bool) propGap {
		return propGap{name: "name", ref: schemaRef{node: node, pointer: widgetPtr + "name"}, required: required, members: []Member{{Field: "name", Pointer: widgetPtr + "name"}}}
	}
	fingerprintAfter := func(order ...bool) string {
		groups := make(map[string]*groupBuild)
		for _, required := range order {
			addGroup(groups, "r", prop(required))
		}
		group, err := groups["r#"+widgetPtr+"name"].finish(doc)
		require.NoError(t, err)
		return group.Fingerprint
	}

	assert.Equal(t, fingerprintAfter(false, true), fingerprintAfter(true, false))
	assert.NotEqual(t, fingerprintAfter(false), fingerprintAfter(true))
}

func TestDiscoverFailsWhenUnionBranchCountDiffersFromGoVariants(t *testing.T) {
	oas := fixtureOASVariant(t, "            - $ref: '#/components/schemas/SourceB'\n",
		"            - $ref: '#/components/schemas/SourceB'\n            - $ref: '#/components/schemas/SourceB'\n")
	_, err := discoverFixture(t, oas)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "oneOf/anyOf branches")
}

func TestDiscoverFailsOnGoFieldWithoutOASProperty(t *testing.T) {
	_, err := discoverFixture(t, fixtureOASVariant(t, "        retries: {type: integer}\n", ""))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "retries")
}

func TestDiscoverFailsWhenAClientOperationIsMissingFromTheOAS(t *testing.T) {
	_, err := discoverFixture(t, fixtureOASVariant(t, "  /api/blobs:", "  /api/blobz:"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing from the OpenAPI document")
}

func TestDiscoverTreatsTypesHandledByReflectionAsCovered(t *testing.T) {
	d := mustDiscover(t)
	for _, resource := range []string{"fx_reflecting", "fx_reflectany"} {
		for _, g := range d.Groups {
			if g.Resource == resource {
				assert.NotContains(t, g.Pointer, "/components/schemas/", g.Key)
			}
		}
	}
}
