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

package kibanaoapi

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// GetPrebuiltRulesStatus retrieves the status of prebuilt rules and timelines for a given space.
func GetPrebuiltRulesStatus(ctx context.Context, client *Client, spaceID string) (*kbapi.ReadPrebuiltRulesAndTimelinesStatusResponse, diag.Diagnostics) {
	resp, err := client.API.ReadPrebuiltRulesAndTimelinesStatusWithResponse(ctx, SpaceAwarePathRequestEditor(spaceID))

	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	if resp.StatusCode() != 200 {
		return nil, diagutil.FrameworkDiagFromError(fmt.Errorf("failed to get prebuilt rules status: %s", resp.Status()))
	}

	return resp, nil
}

// InstallPrebuiltRules installs or updates prebuilt rules and timelines for a given space.
func InstallPrebuiltRules(ctx context.Context, client *Client, spaceID string) diag.Diagnostics {
	resp, err := client.API.InstallPrebuiltRulesAndTimelinesWithResponse(ctx, SpaceAwarePathRequestEditor(spaceID))

	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	if resp.StatusCode() != 200 {
		return diagutil.CheckHTTPErrorFromFW(resp.HTTPResponse, "failed to install prebuilt rules")
	}

	return nil
}
