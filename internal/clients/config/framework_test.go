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

package config

import (
	"context"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

// TestNewFromFrameworkKibanaResource_usesKibanaConnectionOnly verifies that a
// resource-level kibana_connection is the sole source for the scoped client.
// Provider-level fleet blocks cannot reach this path; see also
// Test_newKibanaOapiConfigFromFramework_doesNotApplyFleetFallback.
func TestNewFromFrameworkKibanaResource_usesKibanaConnectionOnly(t *testing.T) {
	os.Unsetenv("KIBANA_ENDPOINT")
	os.Unsetenv("FLEET_ENDPOINT")

	kibanaConns := []KibanaConnection{
		{
			Endpoints: types.ListValueMust(types.StringType, []attr.Value{
				types.StringValue("https://override.example.com"),
			}),
			CACerts:  types.ListValueMust(types.StringType, []attr.Value{}),
			Insecure: types.BoolValue(false),
		},
	}

	client, diags := NewFromFrameworkKibanaResource(context.Background(), kibanaConns, "test")

	require.False(t, diags.HasError())
	require.NotNil(t, client)
	require.NotNil(t, client.KibanaOapi)
	require.Equal(t, "https://override.example.com", client.KibanaOapi.URL)
	require.Equal(t, "https://override.example.com", client.Fleet.URL)
}
