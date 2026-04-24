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

package agentbuilderagent

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/resourcecore"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var (
	_                               resource.Resource                = newAgentResource()
	_                               resource.ResourceWithConfigure   = newAgentResource()
	_                               resource.ResourceWithImportState = newAgentResource()
	minKibanaAgentBuilderAPIVersion                                  = version.Must(version.NewVersion("9.3.0"))
)

type AgentResource struct {
	*resourcecore.Core
}

func newAgentResource() *AgentResource {
	return &AgentResource{
		Core: resourcecore.New(resourcecore.ComponentKibana, "agentbuilder_agent"),
	}
}

// NewResource is a helper function to simplify the provider implementation.
func NewResource() resource.Resource {
	return newAgentResource()
}

func (r *AgentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
