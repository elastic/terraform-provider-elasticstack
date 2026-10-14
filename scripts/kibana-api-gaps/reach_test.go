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
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func reachFor(t *testing.T, pkg string) *reachResult {
	t.Helper()
	return computeReach(loadFixture(t, "testdata/fixture"), fixtureKbapi, "fixture/"+pkg)
}

func sortedMethods(r *reachResult) []string {
	var names []string
	for name := range r.Methods {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func TestReachFindsClientMethodsCalledDirectly(t *testing.T) {
	assert.Equal(t, []string{"CreateWidgetWithResponse"}, sortedMethods(reachFor(t, "direct")))
}

func TestReachFollowsFunctionValuesHandedToACore(t *testing.T) {
	assert.Equal(t, []string{"CreateWidgetWithResponse"}, sortedMethods(reachFor(t, "callback")))
}

func TestReachFollowsCallsIntoSharedHelperPackages(t *testing.T) {
	assert.Equal(t, []string{"CreateWidgetWithResponse"}, sortedMethods(reachFor(t, "helperuser")))
}

func TestReachIsEmptyForResourcesNotCallingTheClient(t *testing.T) {
	assert.Empty(t, reachFor(t, "nokibana").Methods)
}

func TestReachDoesNotExpandIntoProviderClientPlumbing(t *testing.T) {
	// nokibana calls Scoped.Check through an interface; following it would reach GetWidget.
	assert.Empty(t, reachFor(t, "nokibana").Methods)
}

func TestReachResolvesInterfaceCallsToGenericImplementations(t *testing.T) {
	assert.Equal(t, []string{"GetWidgetWithResponse"}, sortedMethods(reachFor(t, "generic")))
}

func TestReachScansAllMethodsOfReferencedModuleTypes(t *testing.T) {
	assert.Equal(t, []string{"CreateWidgetWithResponse"}, sortedMethods(reachFor(t, "embed")))
}
