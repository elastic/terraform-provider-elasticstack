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

package datafeedstate

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/resourcecore"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func NewMLDatafeedStateResource() resource.Resource {
	return &mlDatafeedStateResource{
		Core: resourcecore.New(resourcecore.ComponentElasticsearch, "ml_datafeed_state"),
	}
}

type mlDatafeedStateResource struct {
	*resourcecore.Core
}

func (r *mlDatafeedStateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to datafeed_id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("datafeed_id"), req, resp)
}
