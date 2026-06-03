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

package securitydetectionrule

import (
	"reflect"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/stretchr/testify/require"
)

func TestDetectionRuleModel_satisfiesKibanaResourceModel(t *testing.T) {
	t.Parallel()
	var _ entitycore.KibanaResourceModel = Data{}
}

func TestDetectionRuleModel_satisfiesWithVersionRequirements(t *testing.T) {
	t.Parallel()
	var _ entitycore.WithVersionRequirements = Data{}
}

func TestDetectionRuleResource_embedsEntityCoreKibanaResource(t *testing.T) {
	t.Parallel()
	rt := reflect.TypeFor[securityDetectionRuleResource]()
	field, ok := rt.FieldByName("KibanaResource")
	require.True(t, ok, "securityDetectionRuleResource should embed *entitycore.KibanaResource[Data] as field KibanaResource")
	require.True(t, field.Anonymous)
	require.Equal(t, reflect.TypeFor[*entitycore.KibanaResource[Data]](), field.Type)
}

func TestNewResource_satisfiesFrameworkInterfaces(t *testing.T) {
	t.Parallel()
	var _ resource.Resource = newSecurityDetectionRuleResource()
	var _ resource.ResourceWithConfigure = newSecurityDetectionRuleResource()
	var _ resource.ResourceWithImportState = newSecurityDetectionRuleResource()
	var _ resource.ResourceWithUpgradeState = newSecurityDetectionRuleResource()
}
