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

package provider_test

import (
	"context"
	"strings"
	"testing"

	providerschema "github.com/elastic/terraform-provider-elasticstack/internal/schema"
	"github.com/elastic/terraform-provider-elasticstack/provider"
)

const (
	kbEntityPrefix       = "elasticstack_kibana_"
	fleetEntityPrefix    = "elasticstack_fleet_"
	kbConnectionBlockKey = "kibana_connection"
)

func TestSDKKibanaEntities_ConnectionSchemaMatchesHelper(t *testing.T) {
	p := provider.New("dev")
	expected := providerschema.GetKibanaEntityConnectionSchema()

	runSDKConnectionEntitySubtests(t, "resource", p.ResourcesMap, kbConnectionBlockKey, expected, func(_, name string) bool {
		return strings.HasPrefix(name, kbEntityPrefix)
	})
	runSDKConnectionEntitySubtests(t, "data_source", p.DataSourcesMap, kbConnectionBlockKey, expected, func(_, name string) bool {
		return strings.HasPrefix(name, kbEntityPrefix)
	})
}

func TestSDKFleetEntities_ConnectionSchemaMatchesHelper(t *testing.T) {
	p := provider.New("dev")
	expected := providerschema.GetKibanaEntityConnectionSchema()

	entityCount := 0
	for name := range p.ResourcesMap {
		if strings.HasPrefix(name, fleetEntityPrefix) {
			entityCount++
		}
	}
	for name := range p.DataSourcesMap {
		if strings.HasPrefix(name, fleetEntityPrefix) {
			entityCount++
		}
	}
	if entityCount == 0 {
		t.Skip("no SDK fleet entities registered — kept as a future safety net")
	}

	runSDKConnectionEntitySubtests(t, "resource", p.ResourcesMap, kbConnectionBlockKey, expected, func(_, name string) bool {
		return strings.HasPrefix(name, fleetEntityPrefix)
	})
	runSDKConnectionEntitySubtests(t, "data_source", p.DataSourcesMap, kbConnectionBlockKey, expected, func(_, name string) bool {
		return strings.HasPrefix(name, fleetEntityPrefix)
	})
}

func TestFrameworkKibanaEntities_ConnectionSchemaMatchesHelper(t *testing.T) {
	ctx := context.Background()
	baseProvider := provider.NewFrameworkProvider("dev")
	expected := providerschema.GetKbFWConnectionBlock()

	resourceEntities := collectFrameworkResourceEntities(ctx, baseProvider, func(name string) bool {
		return strings.HasPrefix(name, kbEntityPrefix)
	})
	dataSourceEntities := collectFrameworkDataSourceEntities(ctx, baseProvider, func(name string) bool {
		return strings.HasPrefix(name, kbEntityPrefix)
	})

	runFrameworkConnectionResourceSubtests(ctx, t, resourceEntities, kbConnectionBlockKey, expected)
	runFrameworkConnectionDataSourceSubtests(ctx, t, dataSourceEntities, kbConnectionBlockKey, expected)
}

func TestFrameworkFleetEntities_ConnectionSchemaMatchesHelper(t *testing.T) {
	ctx := context.Background()
	baseProvider := provider.NewFrameworkProvider("dev")
	expected := providerschema.GetKbFWConnectionBlock()

	resourceEntities := collectFrameworkResourceEntities(ctx, baseProvider, func(name string) bool {
		return strings.HasPrefix(name, fleetEntityPrefix)
	})
	dataSourceEntities := collectFrameworkDataSourceEntities(ctx, baseProvider, func(name string) bool {
		return strings.HasPrefix(name, fleetEntityPrefix)
	})

	runFrameworkConnectionResourceSubtests(ctx, t, resourceEntities, kbConnectionBlockKey, expected)
	runFrameworkConnectionDataSourceSubtests(ctx, t, dataSourceEntities, kbConnectionBlockKey, expected)
}
