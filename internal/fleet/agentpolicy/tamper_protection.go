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

package agentpolicy

import (
	"context"
	"fmt"
	"time"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/asyncutils"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
)

const tamperProtectionNotEnabledDetail = "Tamper protection can only be enabled when an Elastic Defend integration policy " +
	"is attached to this agent policy. First apply with is_protected = false, attach " +
	"Elastic Defend, then apply again with is_protected = true. Also ensure Elastic " +
	"Stack 8.10.0 or later, that your license allows tamper protection, and that the " +
	"Fleet API accepts is_protected on this deployment."

// waitForTamperProtection polls Fleet until is_protected becomes true or the timeout elapses.
// On success it returns the reloaded policy; on timeout or reload error it returns the last policy and an error.
func waitForTamperProtection(
	ctx context.Context,
	fleetClient *fleet.Client,
	policyID, spaceID string,
	policy *kbapi.KibanaHTTPAPIsAgentPolicyResponse,
) (*kbapi.KibanaHTTPAPIsAgentPolicyResponse, error) {
	waitCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var reloaded *kbapi.KibanaHTTPAPIsAgentPolicyResponse
	waitErr := asyncutils.WaitForStateTransition(waitCtx, "fleet agent policy", policyID, func(waitCtx context.Context) (bool, error) {
		got, getDiags := fleet.GetAgentPolicy(waitCtx, fleetClient, policyID, spaceID)
		if getDiags.HasError() {
			return false, fmt.Errorf("failed to reload agent policy: %v", getDiags)
		}
		if got == nil {
			return false, nil
		}
		if got.IsProtected {
			reloaded = got
			return true, nil
		}
		return false, nil
	})
	if waitErr != nil {
		return policy, waitErr
	}
	if reloaded != nil {
		return reloaded, nil
	}
	return policy, nil
}
