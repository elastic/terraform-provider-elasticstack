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
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

type alertingRuleFeatures struct {
	SupportsFrequency       bool
	SupportsAlertsFilter    bool
	SupportsAlertDelay      bool
	SupportsFlapping        bool
	SupportsFlappingEnabled bool
}

var alertingRuleFeaturesAllSupported = alertingRuleFeatures{
	SupportsFrequency:       true,
	SupportsAlertsFilter:    true,
	SupportsAlertDelay:      true,
	SupportsFlapping:        true,
	SupportsFlappingEnabled: true,
}

func alertingRuleFeaturesFromVersion(v *version.Version) alertingRuleFeatures {
	return alertingRuleFeatures{
		SupportsFrequency:       v.GreaterThanOrEqual(frequencyMinSupportedVersion),
		SupportsAlertsFilter:    v.GreaterThanOrEqual(alertsFilterMinSupportedVersion),
		SupportsAlertDelay:      v.GreaterThanOrEqual(alertDelayMinSupportedVersion),
		SupportsFlapping:        v.GreaterThanOrEqual(flappingMinSupportedVersion),
		SupportsFlappingEnabled: v.GreaterThanOrEqual(flappingEnabledMinSupportedVersion),
	}
}

func resolveAlertingRuleFeatures(ctx context.Context, client *clients.KibanaScopedClient) (alertingRuleFeatures, diag.Diagnostics) {
	var features alertingRuleFeatures
	populated := false

	_, diags := client.EnforceVersionCheck(ctx, func(v *version.Version) bool {
		populated = true
		features = alertingRuleFeaturesFromVersion(v)
		return true
	})
	if diags.HasError() {
		return alertingRuleFeatures{}, diags
	}
	if !populated {
		return alertingRuleFeaturesAllSupported, nil
	}
	return features, nil
}
