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

package integrationpolicy

// Test-only re-exports for the typed-vars conversion helpers so the external
// integrationpolicy_test package can construct PackagePolicy /
// PackagePolicyRequest values with the typed wrapper types generated from the
// Fleet vars union spec.

// VarsMapToTypedMap is the exported test alias for varsMapToTypedMap.
func VarsMapToTypedMap[T any](m map[string]any) *map[string]*T {
	return varsMapToTypedMap[T](m)
}

// VarsMapToUnionWrapper is the exported test alias for varsMapToUnionWrapper.
func VarsMapToUnionWrapper[T any](m map[string]any) *T {
	return varsMapToUnionWrapper[T](m)
}
