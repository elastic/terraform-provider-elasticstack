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

package cloudconnector

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	fleetclient "github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

const forceDeleteHint = "Set force_delete = true to delete anyway. " +
	"Note: this is destructive and will leave the package policies broken."

var (
	packagePolicyCountPattern   = regexp.MustCompile(`"package_policy_count"\s*:\s*(\d+)`)
	packagePoliciesInUsePattern = regexp.MustCompile(`used by (\d+) package policies`)
)

func deleteCloudConnector(ctx context.Context, client *clients.KibanaScopedClient, resourceID, spaceID string, model cloudConnectorModel) diag.Diagnostics {
	fleetClient := client.GetFleetClient()

	force := model.ForceDelete.ValueBool()
	diags := fleetclient.DeleteCloudConnector(ctx, fleetClient, spaceID, resourceID, force)
	if diags.HasError() && !force {
		return augmentInUseConflictDiagnostic(diags)
	}
	return diags
}

// augmentInUseConflictDiagnostic appends a force_delete hint when the API error
// body mentions package_policy_count. When the count can be parsed, it is included
// in the supplemental diagnostic.
func augmentInUseConflictDiagnostic(diags diag.Diagnostics) diag.Diagnostics {
	for _, d := range diags {
		if d.Severity() != diag.SeverityError {
			continue
		}
		if !strings.Contains(d.Detail(), "package_policy_count") &&
			!strings.Contains(d.Detail(), "package policies") &&
			!strings.Contains(d.Summary(), "package policies") {
			continue
		}

		count, ok := packagePolicyCountFromDetail(d.Detail())
		if !ok {
			count, ok = packagePolicyCountFromDetail(d.Summary())
		}
		if !ok {
			if matches := packagePoliciesInUsePattern.FindStringSubmatch(d.Detail()); len(matches) == 2 {
				if parsed, err := strconv.ParseInt(matches[1], 10, 64); err == nil {
					count, ok = parsed, true
				}
			}
		}
		if !ok {
			if matches := packagePoliciesInUsePattern.FindStringSubmatch(d.Summary()); len(matches) == 2 {
				if parsed, err := strconv.ParseInt(matches[1], 10, 64); err == nil {
					count, ok = parsed, true
				}
			}
		}

		if ok {
			policyWord := "policies"
			if count == 1 {
				policyWord = "policy"
			}
			diags.AddError(
				"Cloud connector in use",
				fmt.Sprintf(
					"This cloud connector is referenced by %d package %s. %s",
					count,
					policyWord,
					forceDeleteHint,
				),
			)
		} else {
			diags.AddError("Cloud connector in use", forceDeleteHint)
		}
		break
	}
	return diags
}

func packagePolicyCountFromDetail(detail string) (int64, bool) {
	var payload struct {
		PackagePolicyCount json.Number `json:"package_policy_count"`
	}
	if err := json.Unmarshal([]byte(detail), &payload); err == nil {
		if count, err := payload.PackagePolicyCount.Int64(); err == nil {
			return count, true
		}
	}

	matches := packagePolicyCountPattern.FindStringSubmatch(detail)
	if len(matches) != 2 {
		return 0, false
	}
	count, err := strconv.ParseInt(matches[1], 10, 64)
	if err != nil {
		return 0, false
	}
	return count, true
}
