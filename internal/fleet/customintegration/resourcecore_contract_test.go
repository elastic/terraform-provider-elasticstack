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
	"reflect"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/resourcecore"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/stretchr/testify/require"
)

func TestCustomIntegrationResource_embedsResourceCore(t *testing.T) {
	t.Parallel()
	r := newCustomIntegrationResource()
	rt := reflect.TypeOf(r).Elem()
	field, ok := rt.FieldByName("Core")
	require.True(t, ok)
	require.True(t, field.Anonymous)
	require.Equal(t, reflect.TypeFor[*resourcecore.Core](), field.Type)
}

func TestCustomIntegrationResource_noImportSupport(t *testing.T) {
	t.Parallel()
	r := NewResource()
	_, ok := any(r).(resource.ResourceWithImportState)
	require.False(t, ok, "custom integration has no ImportState")
}
