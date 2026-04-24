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

package monitor

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/synthetics"
	"github.com/elastic/terraform-provider-elasticstack/internal/resourcecore"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Resource represents a synthetics monitor resource
type Resource struct {
	*resourcecore.Core
}

func newResource() *Resource {
	return &Resource{
		Core: resourcecore.New(resourcecore.ComponentKibana, "synthetics_monitor"),
	}
}

// Ensure provider-defined types fully satisfy framework interfaces
var (
	_ resource.Resource                     = newResource()
	_ resource.ResourceWithConfigure        = newResource()
	_ resource.ResourceWithImportState      = newResource()
	_ resource.ResourceWithConfigValidators = newResource()
	_ synthetics.ESAPIClient                = newResource()
)

// NewResource creates a new synthetics monitor resource
func NewResource() resource.Resource {
	return newResource()
}

func (r *Resource) GetClient() *clients.KibanaScopedClient {
	if r.Client() == nil {
		return nil
	}
	return clients.NewKibanaScopedClientFromFactory(r.Client())
}

func (r *Resource) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.ExactlyOneOf(
			path.MatchRoot("http"),
			path.MatchRoot("tcp"),
			path.MatchRoot("icmp"),
			path.MatchRoot("browser"),
		),
		resourcevalidator.AtLeastOneOf(
			path.MatchRoot("locations"),
			path.MatchRoot("private_locations"),
		),
	}
}

func (r *Resource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *Resource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = monitorConfigSchema()
}
