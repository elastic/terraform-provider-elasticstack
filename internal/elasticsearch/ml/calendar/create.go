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
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	fwdiags "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func createCalendar(ctx context.Context, client *clients.ElasticsearchScopedClient, resourceID string, plan TFModel) (TFModel, fwdiags.Diagnostics) {
	var diags fwdiags.Diagnostics

	calendarID := resourceID

	tflog.Debug(ctx, fmt.Sprintf("Creating ML calendar: %s", calendarID))

	typedClient, err := client.GetESClient()
	if err != nil {
		diags.AddError("Failed to get Elasticsearch client", err.Error())
		return plan, diags
	}

	_, err = typedClient.Ml.PutCalendar(calendarID).Request(newPutCalendarRequestFromTFModel(plan)).Do(ctx)
	if err != nil {
		diags.AddError("Failed to create ML calendar", fmt.Sprintf("Unable to create ML calendar: %s — %s", calendarID, err.Error()))
		return plan, diags
	}

	compID, sdkDiags := client.ID(ctx, calendarID)
	diags.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if diags.HasError() {
		return plan, diags
	}

	plan.ID = types.StringValue(compID.String())

	tflog.Debug(ctx, fmt.Sprintf("Successfully created ML calendar: %s", calendarID))
	return plan, diags
}
