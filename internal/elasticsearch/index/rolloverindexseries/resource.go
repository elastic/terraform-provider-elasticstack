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

package rolloverindexseries

import (
	"context"
	"fmt"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var (
	_ resource.Resource                = newRolloverIndexSeriesResource()
	_ resource.ResourceWithConfigure   = newRolloverIndexSeriesResource()
	_ resource.ResourceWithImportState = newRolloverIndexSeriesResource()
)

type rolloverIndexSeriesResource struct {
	*entitycore.ElasticsearchResource[Model]
}

func newRolloverIndexSeriesResource() *rolloverIndexSeriesResource {
	return &rolloverIndexSeriesResource{
		ElasticsearchResource: entitycore.NewElasticsearchResource[Model]("rollover_index_series", entitycore.ElasticsearchResourceOptions[Model]{
			Schema: getSchema,
			Read:   readRolloverIndexSeries,
			Delete: deleteRolloverIndexSeries,
			Create: createRolloverIndexSeries,
			Update: updateRolloverIndexSeries,
		}),
	}
}

// NewRolloverIndexSeriesResource returns the rollover index series resource for provider registration.
func NewRolloverIndexSeriesResource() resource.Resource {
	return newRolloverIndexSeriesResource()
}

// ImportState accepts <cluster_uuid>/<rollover_alias>/<prefix> and assigns the
// canonical two-part id plus both names and deletion protection before Read.
// ParseCompositeID only splits the cluster segment; it does not set state.
func (r *rolloverIndexSeriesResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	compID, ok := entitycore.ParseCompositeID(req, resp)
	if !ok {
		invalidImportID(resp, req.ID, "is missing a separator")
		return
	}
	if compID.ClusterID == "" {
		invalidImportID(resp, req.ID, "has an empty cluster UUID")
		return
	}
	alias, prefix, found := strings.Cut(compID.ResourceID, "/")
	if !found {
		invalidImportID(resp, req.ID, "is missing a segment")
		return
	}
	if alias == "" || prefix == "" {
		invalidImportID(resp, req.ID, "has an empty segment")
		return
	}
	if strings.Contains(prefix, "/") {
		invalidImportID(resp, req.ID, "has an extra segment")
		return
	}
	if err := ValidateNamingFields(prefix, alias); err != nil {
		invalidImportID(resp, req.ID, "has an invalid naming field: "+err.Error())
		return
	}
	canonical := (&clients.CompositeID{ClusterID: compID.ClusterID, ResourceID: alias}).String()
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), canonical)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("rollover_alias"), alias)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("prefix"), prefix)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("deletion_protection"), true)...)
}

func invalidImportID(resp *resource.ImportStateResponse, id, problem string) {
	resp.Diagnostics.AddError(
		"Invalid import identifier",
		fmt.Sprintf("Import ID %q %s. Expected <cluster_uuid>/<rollover_alias>/<prefix>.", id, problem),
	)
}
