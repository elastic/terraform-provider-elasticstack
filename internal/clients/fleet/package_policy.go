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

package fleet

import (
	"context"
	"net/http"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanautil"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// GetPackagePolicy reads a specific package policy from the API.
func GetPackagePolicy(ctx context.Context, client *Client, id string, spaceID string) (*kbapi.PackagePolicy, diag.Diagnostics) {
	params := kbapi.GetFleetPackagePoliciesPackagepolicyidParams{
		Format: new(kbapi.GetFleetPackagePoliciesPackagepolicyidParamsFormatSimplified),
	}

	resp, err := client.API.GetFleetPackagePoliciesPackagepolicyidWithResponse(ctx, id, &params, kibanautil.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return &resp.JSON200.Item, nil
	case http.StatusNotFound:
		return nil, nil
	default:
		return nil, diagutil.ReportUnknownHTTPError(resp.StatusCode(), resp.Body)
	}
}

// GetDefendPackagePolicy reads a specific Elastic Defend package policy from
// the Fleet API without requesting the simplified format. This preserves the
// typed input shape, input config payloads, and the top-level version token
// required for subsequent update operations.
func GetDefendPackagePolicy(ctx context.Context, client *Client, id string, spaceID string) (*kbapi.PackagePolicy, diag.Diagnostics) {
	resp, err := client.API.GetFleetPackagePoliciesPackagepolicyidWithResponse(ctx, id, nil, kibanautil.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return &resp.JSON200.Item, nil
	case http.StatusNotFound:
		return nil, nil
	default:
		return nil, diagutil.ReportUnknownHTTPError(resp.StatusCode(), resp.Body)
	}
}

// CreatePackagePolicy creates a new package policy.
func CreatePackagePolicy(ctx context.Context, client *Client, spaceID string, req kbapi.PackagePolicyRequest) (*kbapi.PackagePolicy, diag.Diagnostics) {
	params := kbapi.PostFleetPackagePoliciesParams{
		Format: new(kbapi.PostFleetPackagePoliciesParamsFormatSimplified),
	}

	resp, err := client.API.PostFleetPackagePoliciesWithResponse(ctx, &params, req, kibanautil.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return &resp.JSON200.Item, nil
	default:
		return nil, diagutil.ReportUnknownHTTPError(resp.StatusCode(), resp.Body)
	}
}

// CreateDefendPackagePolicy creates a new Elastic Defend package policy using
// the typed-input request body without requesting the simplified format. This
// is used for the Defend bootstrap create step.
func CreateDefendPackagePolicy(ctx context.Context, client *Client, spaceID string, req kbapi.PackagePolicyRequestTypedInputs) (*kbapi.PackagePolicy, diag.Diagnostics) {
	var unionReq kbapi.PackagePolicyRequest
	if err := unionReq.FromPackagePolicyRequestTypedInputs(req); err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	resp, err := client.API.PostFleetPackagePoliciesWithResponse(ctx, nil, unionReq, kibanautil.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return &resp.JSON200.Item, nil
	default:
		return nil, diagutil.ReportUnknownHTTPError(resp.StatusCode(), resp.Body)
	}
}

// UpdatePackagePolicy updates an existing package policy.
func UpdatePackagePolicy(ctx context.Context, client *Client, id string, spaceID string, req kbapi.PackagePolicyRequest) (*kbapi.PackagePolicy, diag.Diagnostics) {
	params := kbapi.PutFleetPackagePoliciesPackagepolicyidParams{
		Format: new(kbapi.Simplified),
	}

	resp, err := client.API.PutFleetPackagePoliciesPackagepolicyidWithResponse(ctx, id, &params, req, kibanautil.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return &resp.JSON200.Item, nil
	default:
		return nil, diagutil.ReportUnknownHTTPError(resp.StatusCode(), resp.Body)
	}
}

// UpdateDefendPackagePolicy updates an existing Elastic Defend package policy
// using the typed-input request body without requesting the simplified format.
// The request body must include the top-level "version" token from the last
// successful read so Kibana can perform optimistic concurrency control.
func UpdateDefendPackagePolicy(ctx context.Context, client *Client, id string, spaceID string, req kbapi.PackagePolicyRequestTypedInputs) (*kbapi.PackagePolicy, diag.Diagnostics) {
	var unionReq kbapi.PackagePolicyRequest
	if err := unionReq.FromPackagePolicyRequestTypedInputs(req); err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	resp, err := client.API.PutFleetPackagePoliciesPackagepolicyidWithResponse(ctx, id, nil, unionReq, kibanautil.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return &resp.JSON200.Item, nil
	default:
		return nil, diagutil.ReportUnknownHTTPError(resp.StatusCode(), resp.Body)
	}
}

// DeletePackagePolicy deletes an existing package policy.
func DeletePackagePolicy(ctx context.Context, client *Client, id string, spaceID string, force bool) diag.Diagnostics {
	params := kbapi.DeleteFleetPackagePoliciesPackagepolicyidParams{
		Force: &force,
	}

	resp, err := client.API.DeleteFleetPackagePoliciesPackagepolicyidWithResponse(ctx, id, &params, kibanautil.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	return handleDeleteResponse(resp.StatusCode(), resp.Body)
}
