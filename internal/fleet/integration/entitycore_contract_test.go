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

package integration

import (
	"reflect"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/stretchr/testify/require"
)

func TestIntegrationModel_satisfiesKibanaResourceModel(t *testing.T) {
	t.Parallel()
	var _ entitycore.KibanaResourceModel = integrationModel{}
	var _ entitycore.KibanaUnscopedSpace = integrationModel{}
}

func TestIntegrationResource_embedsEntityCoreKibanaResource(t *testing.T) {
	t.Parallel()
	rt := reflect.TypeFor[integrationResource]()
	field, ok := rt.FieldByName("KibanaResource")
	require.True(t, ok, "integrationResource should embed *entitycore.KibanaResource[integrationModel] as field KibanaResource")
	require.True(t, field.Anonymous)
	require.Equal(t, reflect.TypeFor[*entitycore.KibanaResource[integrationModel]](), field.Type)
}

func TestNewIntegrationResource_satisfiesFrameworkInterfaces(t *testing.T) {
	t.Parallel()
	var _ resource.Resource = newIntegrationResource()
	var _ resource.ResourceWithConfigure = newIntegrationResource()
	var _ resource.ResourceWithUpgradeState = newIntegrationResource()
}
