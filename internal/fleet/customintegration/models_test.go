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

package customintegration

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

// TestCustomIntegrationModel_GetResourceID_pinsSlashJoinedFormat locks in the
// current "name/version" resource ID format so the
// typeutils.SlashJoinedPackageID consolidation cannot silently change
// existing state IDs.
func TestCustomIntegrationModel_GetResourceID_pinsSlashJoinedFormat(t *testing.T) {
	t.Parallel()

	m := customIntegrationModel{
		PackageName:    types.StringValue("nginx"),
		PackageVersion: types.StringValue("1.2.3"),
	}

	require.Equal(t, "nginx/1.2.3", m.GetResourceID().ValueString())
}

func TestCustomIntegrationModel_GetResourceID_fallsBackToIDWhenUnknown(t *testing.T) {
	t.Parallel()

	m := customIntegrationModel{
		ID:             types.StringValue("fallback-id"),
		PackageName:    types.StringUnknown(),
		PackageVersion: types.StringValue("1.2.3"),
	}

	require.Equal(t, "fallback-id", m.GetResourceID().ValueString())
}
