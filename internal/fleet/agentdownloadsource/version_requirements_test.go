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

package agentdownloadsource

import (
	"context"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

type versionRequirementsTestClient struct {
	version *version.Version
}

func (c versionRequirementsTestClient) EnforceMinVersion(_ context.Context, minVersion *version.Version) (bool, diag.Diagnostics) {
	return !c.version.LessThan(minVersion), nil
}

func sslPopulatedModel(ctx context.Context) model {
	return model{
		Name: types.StringValue("download-source"),
		Host: types.StringValue("https://artifacts.example.com/elastic-agent"),
		Ssl: types.ObjectValueMust(getSslAttrTypes(ctx), map[string]attr.Value{
			"certificate":             types.StringValue("cert"),
			"key":                     types.StringValue("key"),
			"certificate_authorities": types.ListNull(types.StringType),
		}),
	}
}

func TestModelSslVersionRequirementDiagnosticBelow910(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := versionRequirementsTestClient{version: version.Must(version.NewVersion("9.0.0"))}

	diags := entitycore.EnforceVersionRequirements(ctx, client, sslPopulatedModel(ctx))
	require.True(t, diags.HasError())

	var sslDiags []diag.DiagnosticWithPath
	for _, d := range diags {
		if withPath, ok := d.(diag.DiagnosticWithPath); ok && withPath.Path().String() == path.Root("ssl").String() {
			sslDiags = append(sslDiags, withPath)
		}
	}
	require.Len(t, sslDiags, 1, "expected exactly one attribute-scoped diagnostic on ssl")
	require.Contains(t, sslDiags[0].Detail(), "ssl is only supported in Elastic Stack 9.1.0 and above")
}

func TestModelSslVersionRequirementDiagnosticAt910(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := versionRequirementsTestClient{version: version.Must(version.NewVersion("9.1.0"))}

	diags := entitycore.EnforceVersionRequirements(ctx, client, sslPopulatedModel(ctx))
	require.False(t, diags.HasError())
}

func TestModelSslOmittedOnOldVersionHasNoSslDiagnostic(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := versionRequirementsTestClient{version: version.Must(version.NewVersion("9.0.0"))}

	m := sslPopulatedModel(ctx)
	m.Ssl = types.ObjectNull(getSslAttrTypes(ctx))

	diags := entitycore.EnforceVersionRequirements(ctx, client, m)
	require.False(t, diags.HasError(), "ssl omitted on 9.0.0 satisfies every requirement, so no diagnostic should fire")
	for _, d := range diags {
		if withPath, ok := d.(diag.DiagnosticWithPath); ok {
			require.NotEqual(t, path.Root("ssl").String(), withPath.Path().String(), "no ssl diagnostic expected when ssl is not configured")
		}
	}
}
