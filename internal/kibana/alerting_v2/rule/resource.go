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

package rule

import (
	"context"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var (
	_ resource.Resource                = newResource()
	_ resource.ResourceWithConfigure   = newResource()
	_ resource.ResourceWithImportState = newResource()
	_ resource.ResourceWithModifyPlan  = newResource()
)

type Resource struct {
	*entitycore.KibanaResource[Model]
	*entitycore.KibanaSpaceImporter
}

func newResource() *Resource {
	return &Resource{
		KibanaResource: entitycore.NewKibanaResource[Model](
			entitycore.ComponentKibana,
			"alerting_v2_rule",
			entitycore.KibanaResourceOptions[Model]{
				Schema: getSchema,
				Read:   readRule,
				Delete: entitycore.SimpleKibanaDelete[Model](kibanaoapi.DeleteAlertingV2Rule),
				Create: createRule,
				Update: updateRule,
			},
		),
		KibanaSpaceImporter: entitycore.NewKibanaSpaceImporter(path.Root("id"), path.Root("space_id"), path.Root("rule_id")),
	}
}

// NewResource is a helper function to simplify the provider implementation.
func NewResource() resource.Resource {
	return newResource()
}

// ImportState requires an id of the form <space_id>/<rule_id> with both
// segments present. The embedded importer alone would accept "/my-rule" with
// an empty space and report a cluster-oriented format for ids without a
// separator, so the id is validated here before any attribute is seeded.
func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	composite, ok := parseImportID(req.ID)
	if !ok {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			"The import ID must be in the format \"<space_id>/<rule_id>\", for example \"default/my-rule\". Got: \""+req.ID+"\".",
		)
		return
	}
	r.SeedState(ctx, resp, req.ID, composite)
}

func parseImportID(importID string) (*clients.CompositeID, bool) {
	spaceID, ruleID, found := strings.Cut(importID, "/")
	if !found || spaceID == "" || ruleID == "" || strings.Contains(ruleID, "/") {
		return nil, false
	}
	return &clients.CompositeID{ClusterID: spaceID, ResourceID: ruleID}, true
}
