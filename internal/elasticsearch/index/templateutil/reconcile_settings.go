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

package templateutil

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// ReconcileSettingsIfSemanticallyEqual returns stateSettings when planSettings and stateSettings
// differ only in key encoding (dotted vs nested) but are semantically equal. Callers should
// replace plan settings with the returned value when changed is true so Terraform shows no diff.
func ReconcileSettingsIfSemanticallyEqual(ctx context.Context, planSettings, stateSettings customtypes.IndexSettingsValue) (customtypes.IndexSettingsValue, bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	if planSettings.Equal(stateSettings) {
		return planSettings, false, diags
	}

	eq, d := stateSettings.SemanticallyEqual(ctx, planSettings)
	diags.Append(d...)
	if diags.HasError() {
		return planSettings, false, diags
	}
	if eq {
		return stateSettings, true, diags
	}

	return planSettings, false, diags
}
