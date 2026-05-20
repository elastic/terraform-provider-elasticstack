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

package agentbuilderskill

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

const defaultSpaceID = "default"

var (
	_                                     resource.Resource                = newSkillResource()
	_                                     resource.ResourceWithConfigure   = newSkillResource()
	_                                     resource.ResourceWithImportState = newSkillResource()
	minKibanaAgentBuilderSkillsAPIVersion                                  = version.Must(version.NewVersion("9.4.0-SNAPSHOT"))
)

type SkillResource struct {
	*entitycore.ResourceBase
}

func newSkillResource() *SkillResource {
	return &SkillResource{
		ResourceBase: entitycore.NewResourceBase(entitycore.ComponentKibana, "agentbuilder_skill"),
	}
}

// NewResource is a helper function to simplify the provider implementation.
func NewResource() resource.Resource {
	return newSkillResource()
}

func (r *SkillResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
