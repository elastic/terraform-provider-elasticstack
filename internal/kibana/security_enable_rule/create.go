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

package securityenablerule

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func createSecurityEnableRule(
	ctx context.Context,
	client *clients.KibanaScopedClient,
	req entitycore.KibanaWriteRequest[enableRuleModel],
) (entitycore.KibanaWriteResult[enableRuleModel], diag.Diagnostics) {
	return writeSecurityEnableRule(ctx, client, req)
}

func writeSecurityEnableRule(
	ctx context.Context,
	client *clients.KibanaScopedClient,
	req entitycore.KibanaWriteRequest[enableRuleModel],
) (entitycore.KibanaWriteResult[enableRuleModel], diag.Diagnostics) {
	model := req.Plan
	var diags diag.Diagnostics

	oapiClient, err := client.GetKibanaOapiClient()
	if err != nil {
		diags.AddError(err.Error(), "Failed to get Kibana client")
		return entitycore.KibanaWriteResult[enableRuleModel]{}, diags
	}

	spaceID := model.SpaceID.ValueString()
	key := model.Key.ValueString()
	value := model.Value.ValueString()

	if model.DisableOnDestroy.IsNull() {
		model.DisableOnDestroy = types.BoolValue(true)
	}

	model.ID = types.StringValue(fmt.Sprintf("%s/%s:%s", spaceID, key, value))

	diags.Append(kibanaoapi.EnableRulesByTag(ctx, oapiClient, spaceID, key, value)...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[enableRuleModel]{}, diags
	}

	model.AllRulesEnabled = types.BoolValue(true)

	return entitycore.KibanaWriteResult[enableRuleModel]{Model: model}, diags
}
