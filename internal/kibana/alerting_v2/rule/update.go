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

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func updateRule(ctx context.Context, client *clients.KibanaScopedClient, req entitycore.KibanaWriteRequest[Model]) (entitycore.KibanaWriteResult[Model], diag.Diagnostics) {
	created, written, diags := putRule(ctx, client, req.SpaceID, req.WriteID, req.Plan)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[Model]{}, diags
	}
	if created {
		importID := clients.CompositeIDValue(req.SpaceID, req.WriteID).ValueString()
		diags.AddError(
			"Alerting v2 rule was recreated",
			"The rule \""+importID+"\" was deleted outside Terraform after the plan was produced, and this update recreated it with a new created_at. "+
				"Run a new plan to reconcile the state with the recreated rule.",
		)
		return entitycore.KibanaWriteResult[Model]{}, diags
	}

	return entitycore.KibanaWriteResult[Model]{Model: written}, diags
}
