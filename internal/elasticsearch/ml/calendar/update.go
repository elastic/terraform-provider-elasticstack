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

package calendar

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	fwdiags "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func updateCalendar(ctx context.Context, client *clients.ElasticsearchScopedClient, resourceID string, plan TFModel, prior TFModel) (TFModel, fwdiags.Diagnostics) {
	var diags fwdiags.Diagnostics

	calendarID := resourceID
	if calendarID == "" {
		diags.AddError("Invalid resource ID", "calendar_id cannot be empty")
		return plan, diags
	}

	tflog.Debug(ctx, fmt.Sprintf("Updating ML calendar: %s", calendarID))

	typedClient, err := client.GetESClient()
	if err != nil {
		diags.AddError("Failed to get Elasticsearch client", err.Error())
		return plan, diags
	}

	putModel := plan
	if !typeutils.IsKnown(plan.Description) && typeutils.IsKnown(prior.Description) {
		putModel.Description = prior.Description
	}

	_, err = typedClient.Ml.PutCalendar(calendarID).Request(newPutCalendarRequestFromTFModel(putModel)).Do(ctx)
	if err != nil {
		diags.AddError("Failed to update ML calendar", fmt.Sprintf("Unable to update ML calendar: %s — %s", calendarID, err.Error()))
		return plan, diags
	}

	tflog.Debug(ctx, fmt.Sprintf("Successfully updated ML calendar: %s", calendarID))
	return plan, diags
}
