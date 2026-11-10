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
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func readRule(ctx context.Context, client *clients.KibanaScopedClient, resourceID, spaceID string, model Model) (Model, bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	rule, getDiags := kibanaoapi.GetAlertingV2Rule(ctx, client.GetKibanaOapiClient(), spaceID, resourceID)
	diags.Append(withCompositeID(getDiags, spaceID, resourceID)...)
	if diags.HasError() {
		return model, false, diags
	}
	if rule == nil {
		return model, false, diags
	}

	diags.Append(model.populateFromAPI(ctx, rule)...)
	if diags.HasError() {
		return model, false, diags
	}

	model.ID, model.SpaceID = clients.CompositeIdentityValues(spaceID, resourceID)
	model.RuleID = types.StringValue(resourceID)

	return model, true, diags
}

// withCompositeID prefixes every error detail with the rule's composite id so
// that a failed read-after-write still tells the user which id to import: the
// envelope discards state on that failure, and a generated rule_id would
// otherwise be lost.
func withCompositeID(in diag.Diagnostics, spaceID, resourceID string) diag.Diagnostics {
	importID := clients.CompositeIDValue(spaceID, resourceID).ValueString()
	out := make(diag.Diagnostics, 0, len(in))
	for _, d := range in {
		if d.Severity() != diag.SeverityError {
			out = append(out, d)
			continue
		}
		out = append(out, diag.NewErrorDiagnostic(d.Summary(), "Rule \""+importID+"\": "+d.Detail()))
	}
	return out
}
