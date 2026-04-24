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

package privatelocation

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/synthetics"
	"github.com/elastic/terraform-provider-elasticstack/internal/resourcecore"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const resourceName = synthetics.MetadataPrefix + "private_location"

type Resource struct {
	*resourcecore.Core
}

func newResource() *Resource {
	return &Resource{
		Core: resourcecore.New(resourcecore.ComponentKibana, "synthetics_private_location"),
	}
}

// Ensure provider defined types fully satisfy framework interfaces
var (
	_ resource.Resource                = newResource()
	_ resource.ResourceWithConfigure   = newResource()
	_ resource.ResourceWithImportState = newResource()
	_ synthetics.ESAPIClient           = newResource()
)

func (r *Resource) GetClient() *clients.KibanaScopedClient {
	if r.Client() == nil {
		return nil
	}
	return clients.NewKibanaScopedClientFromFactory(r.Client())
}

func (r *Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = privateLocationSchema()
}

func (r *Resource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

// NewResource returns a synthetics private location resource with shared bootstrap wiring.
func NewResource() resource.Resource {
	return newResource()
}

func (r *Resource) Update(ctx context.Context, _ resource.UpdateRequest, response *resource.UpdateResponse) {
	tflog.Warn(ctx, "Update isn't supported for elasticstack_"+resourceName)
	response.Diagnostics.AddError(
		"synthetics private location update not supported",
		"Synthetics private location could only be replaced. Please, note, that only unused locations could be deleted.",
	)
}
