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

package spacesettings

import (
	"context"
	"reflect"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/stretchr/testify/require"
)

func TestResource_embedsKibanaResource(t *testing.T) {
	t.Parallel()
	rt := reflect.TypeFor[Resource]()
	field, ok := rt.FieldByName("KibanaResource")
	require.True(t, ok)
	require.True(t, field.Anonymous)
	require.Equal(t, reflect.TypeFor[*entitycore.KibanaResource[spaceSettingsModel]](), field.Type)
}

func TestResource_typeName(t *testing.T) {
	t.Parallel()

	resp := &resource.MetadataResponse{}
	NewResource().Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "elasticstack"}, resp)

	require.Equal(t, "elasticstack_fleet_space_settings", resp.TypeName)
}

func TestResource_schemaIncludesKibanaConnectionBlock(t *testing.T) {
	t.Parallel()

	resp := &resource.SchemaResponse{}
	NewResource().Schema(context.Background(), resource.SchemaRequest{}, resp)

	require.False(t, resp.Diagnostics.HasError())
	require.Contains(t, resp.Schema.Blocks, "kibana_connection")
	require.Empty(t, resp.Schema.ValidateImplementation(context.Background()))
}
