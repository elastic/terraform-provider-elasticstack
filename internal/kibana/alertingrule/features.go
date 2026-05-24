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

package alertingrule

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

type alertingRuleFeatures struct {
	SupportsFrequency       bool
	SupportsAlertsFilter    bool
	SupportsAlertDelay      bool
	SupportsFlapping        bool
	SupportsFlappingEnabled bool
}

func resolveAlertingRuleFeatures(ctx context.Context, client *clients.KibanaScopedClient) (alertingRuleFeatures, diag.Diagnostics) {
	supportsFrequency, diags := client.EnforceMinVersion(ctx, frequencyMinSupportedVersion)
	if diags.HasError() {
		return alertingRuleFeatures{}, diags
	}

	supportsAlertsFilter, diags := client.EnforceMinVersion(ctx, alertsFilterMinSupportedVersion)
	if diags.HasError() {
		return alertingRuleFeatures{}, diags
	}

	supportsAlertDelay, diags := client.EnforceMinVersion(ctx, alertDelayMinSupportedVersion)
	if diags.HasError() {
		return alertingRuleFeatures{}, diags
	}

	supportsFlapping, diags := client.EnforceMinVersion(ctx, flappingMinSupportedVersion)
	if diags.HasError() {
		return alertingRuleFeatures{}, diags
	}

	supportsFlappingEnabled, diags := client.EnforceMinVersion(ctx, flappingEnabledMinSupportedVersion)
	if diags.HasError() {
		return alertingRuleFeatures{}, diags
	}

	return alertingRuleFeatures{
		SupportsFrequency:       supportsFrequency,
		SupportsAlertsFilter:    supportsAlertsFilter,
		SupportsAlertDelay:      supportsAlertDelay,
		SupportsFlapping:        supportsFlapping,
		SupportsFlappingEnabled: supportsFlappingEnabled,
	}, nil
}
