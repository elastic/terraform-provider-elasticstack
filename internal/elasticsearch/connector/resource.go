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

package connector

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = newContentConnectorResource()
	_ resource.ResourceWithConfigure   = newContentConnectorResource()
	_ resource.ResourceWithImportState = newContentConnectorResource()
	_ resource.ResourceWithModifyPlan  = newContentConnectorResource()
)

type contentConnectorResource struct {
	*entitycore.ElasticsearchResource[ContentConnectorData]
}

func newContentConnectorResource() *contentConnectorResource {
	return &contentConnectorResource{
		ElasticsearchResource: entitycore.NewElasticsearchResource("connector", entitycore.ElasticsearchResourceOptions[ContentConnectorData]{
			Schema: schemaFactory,
			Create: createConnector,
			Read:   readConnector,
			Update: updateConnector,
			Delete: deleteConnector,
		}),
	}
}

// NewContentConnectorResource returns a new content connector resource for registration with the provider.
func NewContentConnectorResource() resource.Resource { return newContentConnectorResource() }

func (r *contentConnectorResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var prior ContentConnectorData
	resp.Diagnostics.Append(req.State.Get(ctx, &prior)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.ElasticsearchResource.Delete(ctx, req, resp)
	if resp.Diagnostics.HasError() {
		return
	}

	clearAllSecretHashesFromPrior(ctx, resp.Private, prior, &resp.Diagnostics)
}
