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

// Package managedintegration implements the elasticstack_fleet_managed_integration
// resource (openspec/changes/fleet-managed-integration). It mirrors the
// structure of internal/fleet/proxy: resource.go wires the entitycore
// Kibana resource envelope, models.go defines the Plugin Framework model,
// schema.go defines the schema, and create.go/read.go/update.go/delete.go
// implement the CRUD callbacks.
//
// As of Task 3 of the OpenSpec change's tasks.md ("3. Resource: skeleton,
// model, and spike"), this package is a skeleton: the schema only carries
// identity attributes and the CRUD callbacks are stubs. The full schema
// lands in Task 4 and full CRUD lands in Task 5.
package managedintegration

import (
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/fleet"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var (
	_ resource.Resource                = newResource()
	_ resource.ResourceWithConfigure   = newResource()
	_ resource.ResourceWithImportState = newResource()
)

// Resource implements the Fleet agentless policy resource.
type Resource struct {
	*entitycore.KibanaResource[agentlessPolicyModel]
	*fleet.SpaceImporter
}

func newResource() *Resource {
	return &Resource{
		KibanaResource: entitycore.NewKibanaResource[agentlessPolicyModel](
			entitycore.ComponentFleet,
			"managed_integration",
			entitycore.KibanaResourceOptions[agentlessPolicyModel]{
				Schema: getSchema,
				Read:   readAgentlessPolicy,
				Delete: deleteAgentlessPolicy,
				Create: createAgentlessPolicy,
				Update: updateAgentlessPolicy,
			},
		),
		SpaceImporter: fleet.NewSpaceImporter(path.Root("policy_id")),
	}
}

// NewResource is a helper function to simplify the provider implementation.
func NewResource() resource.Resource {
	return newResource()
}
