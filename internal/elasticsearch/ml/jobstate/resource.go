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

package jobstate

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var (
	_ resource.Resource                = newMLJobStateResource()
	_ resource.ResourceWithConfigure   = newMLJobStateResource()
	_ resource.ResourceWithImportState = newMLJobStateResource()
)

type mlJobStateResource struct {
	*entitycore.ElasticsearchResource[MLJobStateData]
}

func newMLJobStateResource() *mlJobStateResource {
	createFunc, updateFunc := entitycore.PlaceholderElasticsearchWriteCallbacks[MLJobStateData]()
	return &mlJobStateResource{
		ElasticsearchResource: entitycore.NewElasticsearchResource[MLJobStateData](
			entitycore.ComponentElasticsearch,
			"ml_job_state",
			GetSchema,
			readMLJobState,
			deleteMLJobState,
			createFunc,
			updateFunc,
		),
	}
}

func NewMLJobStateResource() resource.Resource {
	return newMLJobStateResource()
}

func (r *mlJobStateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
