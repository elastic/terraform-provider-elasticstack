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

package slo

import (
	"reflect"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/stretchr/testify/require"
)

func TestTfModel_satisfiesKibanaResourceModel(t *testing.T) {
	t.Parallel()
	var _ entitycore.KibanaResourceModel = tfModel{}
}

func TestResource_embedsEntityCoreKibanaResource(t *testing.T) {
	t.Parallel()
	rt := reflect.TypeFor[Resource]()
	field, ok := rt.FieldByName("KibanaResource")
	require.True(t, ok, "Resource should embed *entitycore.KibanaResource[tfModel] as field KibanaResource")
	require.True(t, field.Anonymous)
	require.Equal(t, reflect.TypeFor[*entitycore.KibanaResource[tfModel]](), field.Type)
}

func TestNewResource_satisfiesFrameworkInterfaces(t *testing.T) {
	t.Parallel()
	var _ resource.Resource = newResource()
	var _ resource.ResourceWithConfigure = newResource()
	var _ resource.ResourceWithImportState = newResource()
	var _ resource.ResourceWithConfigValidators = newResource()
	var _ resource.ResourceWithUpgradeState = newResource()
}
