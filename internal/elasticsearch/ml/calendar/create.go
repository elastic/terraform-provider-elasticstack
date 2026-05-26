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
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	fwdiags "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func createCalendar(ctx context.Context, client *clients.ElasticsearchScopedClient, req entitycore.WriteRequest[TFModel]) (entitycore.WriteResult[TFModel], fwdiags.Diagnostics) {
	var diags fwdiags.Diagnostics
	plan := req.Plan
	calendarID := req.WriteID

	tflog.Debug(ctx, fmt.Sprintf("Creating ML calendar: %s", calendarID))

	typedClient := client.GetESClient()

	_, err := typedClient.Ml.PutCalendar(calendarID).Request(newPutCalendarRequestFromTFModel(plan)).Do(ctx)
	if err != nil {
		diags.AddError("Failed to create ML calendar", fmt.Sprintf("Unable to create ML calendar: %s — %s", calendarID, err.Error()))
		return entitycore.WriteResult[TFModel]{Model: plan}, diags
	}

	compID, idDiags := client.ID(ctx, calendarID)
	diags.Append(idDiags...)
	if diags.HasError() {
		return entitycore.WriteResult[TFModel]{Model: plan}, diags
	}

	plan.ID = types.StringValue(compID.String())

	tflog.Debug(ctx, fmt.Sprintf("Successfully created ML calendar: %s", calendarID))
	return entitycore.WriteResult[TFModel]{Model: plan}, diags
}
