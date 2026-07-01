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

// This file re-exports the Fleet package-policy inputs/streams/vars modeling
// extracted into internal/fleet/policyshape (see that package's doc.go for
// the extraction rationale). Per the Phase 1 refactor's "thin wrapper"
// allowance (openspec/changes/fleet-agentless-policy/tasks.md, task 1.3), the
// aliases below let the rest of this package (models.go, schema.go,
// schema_v1.go, schema_v2.go, create.go, read.go, update.go) keep referring
// to these types/functions by their original, package-local names instead of
// a large mechanical rename to policyshape.Foo everywhere.
//
// Generic functions cannot be aliased via a package-level var (Go does not
// support assigning an uninstantiated generic function to a variable), so
// those get a thin pass-through wrapper instead.

import (
	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/fleet/policyshape"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

type (
	// InputType and InputValue model a single element of the `inputs` map.
	InputType  = policyshape.InputType
	InputValue = policyshape.InputValue

	// InputsType and InputsValue model the top-level `inputs` map attribute.
	InputsType  = policyshape.InputsType
	InputsValue = policyshape.InputsValue

	// VarsJSONType and VarsJSONValue model the top-level `vars_json` attribute.
	VarsJSONType  = policyshape.VarsJSONType
	VarsJSONValue = policyshape.VarsJSONValue

	// integrationPolicyInputsModel/integrationPolicyInputStreamModel are the Go
	// representations of `inputs[key]` and `inputs[key].streams[key]`.
	integrationPolicyInputsModel      = policyshape.InputModel
	integrationPolicyInputStreamModel = policyshape.InputStreamModel

	// inputDefaultsModel is the Go representation of the package-computed
	// `defaults` object nested under an input. Its per-stream counterpart,
	// policyshape.InputDefaultsStreamModel, has no local alias here: it isn't
	// referenced directly by this package (only nested inside
	// policyshape.InputDefaultsModel.Streams).
	inputDefaultsModel = policyshape.InputDefaultsModel
)

var (
	NewInputType       = policyshape.NewInputType
	NewInputsType      = policyshape.NewInputsType
	NewInputsNull      = policyshape.NewInputsNull
	NewInputsValueFrom = policyshape.NewInputsValueFrom

	NewVarsJSONNull    = policyshape.NewVarsJSONNull
	NewVarsJSONUnknown = policyshape.NewVarsJSONUnknown

	inputsConfigured = policyshape.InputsConfigured

	// HandleRespSecrets/HandleReqRespSecrets stay exported under their
	// original names since acc/unit tests in other packages may reference
	// them; see also export_test.go historically. They now delegate to the
	// shared secret-handling implementation.
	HandleRespSecrets    = policyshape.HandleRespSecrets
	HandleReqRespSecrets = policyshape.HandleReqRespSecrets
)

// NewVarsJSONWithIntegration creates a VarsJSONValue with a known value and
// an integration context, using this resource's package-info cache
// (knownPackages, populated by getPackageInfo) to resolve defaults.
func NewVarsJSONWithIntegration(value string, name, version string) (VarsJSONValue, diag.Diagnostics) {
	return policyshape.NewVarsJSONWithIntegration(value, name, version, lookupCachedPackageInfo)
}

// varsAnyToMap and varsMapToTypedMap are generic (or need not be, but are
// kept as thin wrappers for symmetry), so they can't be aliased via a
// package-level var; see the file comment above. (policyshape.VarsMapToUnionWrapper
// has no local wrapper: this package has no caller for it.)
func varsAnyToMap(v any) map[string]any {
	return policyshape.VarsAnyToMap(v)
}

func varsMapToTypedMap[T any](m map[string]any) *map[string]*T {
	return policyshape.VarsMapToTypedMap[T](m)
}

// packageInfoToDefaults derives per-input default values from Fleet package
// metadata (see policyshape.PackageInfoToDefaults).
func packageInfoToDefaults(pkg *kbapi.KibanaHTTPAPIsGetPackageInfo) (map[string]inputDefaultsModel, diag.Diagnostics) {
	return policyshape.PackageInfoToDefaults(pkg)
}
