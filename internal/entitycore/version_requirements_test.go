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

package entitycore

import (
	"context"
	"testing"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/stretchr/testify/require"
)

type versionRequirementsTestModel struct {
	requirements []VersionRequirement
}

func (m versionRequirementsTestModel) GetVersionRequirements(context.Context) ([]VersionRequirement, diag.Diagnostics) {
	return m.requirements, nil
}

type versionRequirementsTestClient struct {
	version *version.Version
}

func (c versionRequirementsTestClient) EnforceMinVersion(_ context.Context, minVersion *version.Version) (bool, diag.Diagnostics) {
	return !c.version.LessThan(minVersion), nil
}

func (c versionRequirementsTestClient) EnforceVersionCheck(_ context.Context, check func(*version.Version) bool) (bool, diag.Diagnostics) {
	return check(c.version), nil
}

func TestEnforceVersionRequirementsAttributePath(t *testing.T) {
	t.Parallel()

	model := versionRequirementsTestModel{requirements: []VersionRequirement{
		NewAttributeVersionRequirement(
			path.Root("feature"),
			*version.Must(version.NewVersion("9.1.0")),
			"feature requires 9.1.0",
		),
	}}

	diags := EnforceVersionRequirements(
		context.Background(),
		versionRequirementsTestClient{version: version.Must(version.NewVersion("9.0.0"))},
		model,
	)
	require.True(t, diags.HasError())
	require.Equal(t, "Unsupported server version", diags[0].Summary())
	withPath, ok := diags[0].(diag.DiagnosticWithPath)
	require.True(t, ok)
	require.Equal(t, path.Root("feature"), withPath.Path())
}

func TestEnforceVersionRequirementsCustomCheck(t *testing.T) {
	t.Parallel()

	check := func(v *version.Version) bool {
		return v.Equal(version.Must(version.NewVersion("8.19.0"))) || !v.LessThan(version.Must(version.NewVersion("9.1.0")))
	}
	model := versionRequirementsTestModel{requirements: []VersionRequirement{
		NewAttributeVersionCheckRequirement(path.Root("feature"), check, "unsupported"),
	}}

	unsupported := EnforceVersionRequirements(
		context.Background(),
		versionRequirementsTestClient{version: version.Must(version.NewVersion("9.0.0"))},
		model,
	)
	require.True(t, unsupported.HasError())

	supported := EnforceVersionRequirements(
		context.Background(),
		versionRequirementsTestClient{version: version.Must(version.NewVersion("9.1.0"))},
		model,
	)
	require.False(t, supported.HasError())
}
