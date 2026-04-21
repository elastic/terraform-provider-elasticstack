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
	"archive/zip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// buildSpaceAwarePath constructs an API path with space awareness.
// If spaceID is empty or "default", returns the basePath unchanged.
// Otherwise, prepends "/s/{spaceID}" to the basePath.
func buildSpaceAwarePath(spaceID, basePath string) string {
	if spaceID != "" && spaceID != "default" {
		return fmt.Sprintf("/s/%s%s", spaceID, basePath)
	}
	return basePath
}

func spaceAwarePathRequestEditor(spaceID string) func(ctx context.Context, req *http.Request) error {
	return func(_ context.Context, req *http.Request) error {
		req.URL.Path = buildSpaceAwarePath(spaceID, req.URL.Path)
		return nil
	}
}

// GetEnrollmentTokens reads all enrollment tokens from the API.
func GetEnrollmentTokens(ctx context.Context, client *Client, spaceID string) ([]kbapi.EnrollmentApiKey, diag.Diagnostics) {
	resp, err := client.API.GetFleetEnrollmentApiKeysWithResponse(ctx, nil, spaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return resp.JSON200.Items, nil
	default:
		return nil, reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// GetEnrollmentTokensByPolicy Get enrollment tokens by given policy ID.
func GetEnrollmentTokensByPolicy(ctx context.Context, client *Client, policyID string) ([]kbapi.EnrollmentApiKey, diag.Diagnostics) {
	params := kbapi.GetFleetEnrollmentApiKeysParams{
		Kuery: new("policy_id:" + policyID),
	}

	resp, err := client.API.GetFleetEnrollmentApiKeysWithResponse(ctx, &params)
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return resp.JSON200.Items, nil
	default:
		return nil, reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// GetEnrollmentTokensByPolicyInSpace Get enrollment tokens by policy ID within a specific Kibana space.
func GetEnrollmentTokensByPolicyInSpace(ctx context.Context, client *Client, policyID string, spaceID string) ([]kbapi.EnrollmentApiKey, diag.Diagnostics) {
	// Construct the space-aware path
	path := buildSpaceAwarePath(spaceID, "/api/fleet/enrollment_api_keys?kuery=policy_id:"+policyID)

	req, err := http.NewRequestWithContext(ctx, "GET", client.URL+path, nil)
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	httpResp, err := client.HTTP.Do(req)
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}
	defer httpResp.Body.Close()

	switch httpResp.StatusCode {
	case http.StatusOK:
		var result struct {
			Items []kbapi.EnrollmentApiKey `json:"items"`
		}
		if err := json.NewDecoder(httpResp.Body).Decode(&result); err != nil {
			return nil, diagutil.FrameworkDiagFromError(err)
		}
		return result.Items, nil
	default:
		bodyBytes, _ := io.ReadAll(httpResp.Body)
		return nil, reportUnknownError(httpResp.StatusCode, bodyBytes)
	}
}

// GetAgentPolicy reads a specific agent policy from the API.
func GetAgentPolicy(ctx context.Context, client *Client, id string, spaceID string) (*kbapi.AgentPolicy, diag.Diagnostics) {
	resp, err := client.API.GetFleetAgentPoliciesAgentpolicyidWithResponse(ctx, id, nil, spaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return &resp.JSON200.Item, nil
	case http.StatusNotFound:
		return nil, nil
	default:
		return nil, reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// CreateAgentPolicy creates a new agent policy.
func CreateAgentPolicy(ctx context.Context, client *Client, req kbapi.PostFleetAgentPoliciesJSONRequestBody, sysMonitoring bool, spaceID string) (*kbapi.AgentPolicy, diag.Diagnostics) {
	params := kbapi.PostFleetAgentPoliciesParams{
		SysMonitoring: new(sysMonitoring),
	}

	resp, err := client.API.PostFleetAgentPoliciesWithResponse(ctx, &params, req, spaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return &resp.JSON200.Item, nil
	default:
		return nil, reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// UpdateAgentPolicy updates an existing agent policy.
func UpdateAgentPolicy(ctx context.Context, client *Client, id string, spaceID string, req kbapi.PutFleetAgentPoliciesAgentpolicyidJSONRequestBody) (*kbapi.AgentPolicy, diag.Diagnostics) {
	resp, err := client.API.PutFleetAgentPoliciesAgentpolicyidWithResponse(ctx, id, nil, req, spaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return &resp.JSON200.Item, nil
	default:
		return nil, reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// DeleteAgentPolicy deletes an existing agent policy.
func DeleteAgentPolicy(ctx context.Context, client *Client, id string, spaceID string) diag.Diagnostics {
	body := kbapi.PostFleetAgentPoliciesDeleteJSONRequestBody{
		AgentPolicyId: id,
	}

	resp, err := client.API.PostFleetAgentPoliciesDeleteWithResponse(ctx, body, spaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return nil
	case http.StatusNotFound:
		return nil
	default:
		return reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// GetOutputs reads all outputs from the API.
func GetOutputs(ctx context.Context, client *Client, spaceID string) ([]kbapi.OutputUnion, diag.Diagnostics) {
	resp, err := client.API.GetFleetOutputsWithResponse(ctx, spaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return resp.JSON200.Items, nil
	case http.StatusNotFound:
		return nil, nil
	default:
		return nil, reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// GetOutput reads a specific output from the API.
func GetOutput(ctx context.Context, client *Client, id string, spaceID string) (*kbapi.OutputUnion, diag.Diagnostics) {
	resp, err := client.API.GetFleetOutputsOutputidWithResponse(ctx, id, spaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return &resp.JSON200.Item, nil
	case http.StatusNotFound:
		return nil, nil
	default:
		return nil, reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// CreateOutput creates a new output.
func CreateOutput(ctx context.Context, client *Client, spaceID string, req kbapi.NewOutputUnion) (*kbapi.OutputUnion, diag.Diagnostics) {
	resp, err := client.API.PostFleetOutputsWithResponse(ctx, req, spaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return &resp.JSON200.Item, nil
	default:
		return nil, reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// UpdateOutput updates an existing output.
func UpdateOutput(ctx context.Context, client *Client, id string, spaceID string, req kbapi.UpdateOutputUnion) (*kbapi.OutputUnion, diag.Diagnostics) {
	resp, err := client.API.PutFleetOutputsOutputidWithResponse(ctx, id, req, spaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return &resp.JSON200.Item, nil
	default:
		return nil, reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// DeleteOutput deletes an existing output.
func DeleteOutput(ctx context.Context, client *Client, id string, spaceID string) diag.Diagnostics {
	resp, err := client.API.DeleteFleetOutputsOutputidWithResponse(ctx, id, spaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return nil
	case http.StatusNotFound:
		return nil
	default:
		return reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// GetFleetServerHost reads a specific fleet server host from the API.
func GetFleetServerHost(ctx context.Context, client *Client, id string, spaceID string) (*kbapi.ServerHost, diag.Diagnostics) {
	resp, err := client.API.GetFleetFleetServerHostsItemidWithResponse(ctx, id, spaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return &resp.JSON200.Item, nil
	case http.StatusNotFound:
		return nil, nil
	default:
		return nil, reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// CreateFleetServerHost creates a new fleet server host.
func CreateFleetServerHost(ctx context.Context, client *Client, spaceID string, req kbapi.PostFleetFleetServerHostsJSONRequestBody) (*kbapi.ServerHost, diag.Diagnostics) {
	resp, err := client.API.PostFleetFleetServerHostsWithResponse(ctx, req, spaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return &resp.JSON200.Item, nil
	default:
		return nil, reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// UpdateFleetServerHost updates an existing fleet server host.
func UpdateFleetServerHost(ctx context.Context, client *Client, id string, spaceID string, req kbapi.PutFleetFleetServerHostsItemidJSONRequestBody) (*kbapi.ServerHost, diag.Diagnostics) {
	resp, err := client.API.PutFleetFleetServerHostsItemidWithResponse(ctx, id, req, spaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return &resp.JSON200.Item, nil
	default:
		return nil, reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// DeleteFleetServerHost deletes an existing fleet server host.
func DeleteFleetServerHost(ctx context.Context, client *Client, id string, spaceID string) diag.Diagnostics {
	resp, err := client.API.DeleteFleetFleetServerHostsItemidWithResponse(ctx, id, spaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return nil
	case http.StatusNotFound:
		return nil
	default:
		return reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// GetPackagePolicy reads a specific package policy from the API.
func GetPackagePolicy(ctx context.Context, client *Client, id string, spaceID string) (*kbapi.PackagePolicy, diag.Diagnostics) {
	params := kbapi.GetFleetPackagePoliciesPackagepolicyidParams{
		Format: new(kbapi.GetFleetPackagePoliciesPackagepolicyidParamsFormatSimplified),
	}

	resp, err := client.API.GetFleetPackagePoliciesPackagepolicyidWithResponse(ctx, id, &params, spaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return &resp.JSON200.Item, nil
	case http.StatusNotFound:
		return nil, nil
	default:
		return nil, reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// GetDefendPackagePolicy reads a specific Elastic Defend package policy from
// the Fleet API without requesting the simplified format. This preserves the
// typed input shape, input config payloads, and the top-level version token
// required for subsequent update operations.
func GetDefendPackagePolicy(ctx context.Context, client *Client, id string, spaceID string) (*kbapi.PackagePolicy, diag.Diagnostics) {
	resp, err := client.API.GetFleetPackagePoliciesPackagepolicyidWithResponse(ctx, id, nil, spaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return &resp.JSON200.Item, nil
	case http.StatusNotFound:
		return nil, nil
	default:
		return nil, reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// CreatePackagePolicy creates a new package policy.
func CreatePackagePolicy(ctx context.Context, client *Client, spaceID string, req kbapi.PackagePolicyRequest) (*kbapi.PackagePolicy, diag.Diagnostics) {
	params := kbapi.PostFleetPackagePoliciesParams{
		Format: new(kbapi.PostFleetPackagePoliciesParamsFormatSimplified),
	}

	resp, err := client.API.PostFleetPackagePoliciesWithResponse(ctx, &params, req, spaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return &resp.JSON200.Item, nil
	default:
		return nil, reportUnknownError(resp.StatusCode(), resp.Body)
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

	resp, err := client.API.PostFleetPackagePoliciesWithResponse(ctx, nil, unionReq, spaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return &resp.JSON200.Item, nil
	default:
		return nil, reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// UpdatePackagePolicy updates an existing package policy.
func UpdatePackagePolicy(ctx context.Context, client *Client, id string, spaceID string, req kbapi.PackagePolicyRequest) (*kbapi.PackagePolicy, diag.Diagnostics) {
	params := kbapi.PutFleetPackagePoliciesPackagepolicyidParams{
		Format: new(kbapi.Simplified),
	}

	resp, err := client.API.PutFleetPackagePoliciesPackagepolicyidWithResponse(ctx, id, &params, req, spaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return &resp.JSON200.Item, nil
	default:
		return nil, reportUnknownError(resp.StatusCode(), resp.Body)
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

	resp, err := client.API.PutFleetPackagePoliciesPackagepolicyidWithResponse(ctx, id, nil, unionReq, spaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return &resp.JSON200.Item, nil
	default:
		return nil, reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// DeletePackagePolicy deletes an existing package policy.
func DeletePackagePolicy(ctx context.Context, client *Client, id string, spaceID string, force bool) diag.Diagnostics {
	params := kbapi.DeleteFleetPackagePoliciesPackagepolicyidParams{
		Force: &force,
	}

	resp, err := client.API.DeleteFleetPackagePoliciesPackagepolicyidWithResponse(ctx, id, &params, spaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return nil
	case http.StatusNotFound:
		return nil
	default:
		return reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// GetAgentDownloadSource reads a specific agent binary download source from the API.
func GetAgentDownloadSource(ctx context.Context, client *Client, id string, spaceID string) (*kbapi.GetFleetAgentDownloadSourcesSourceidResponse, diag.Diagnostics) {
	resp, err := client.API.GetFleetAgentDownloadSourcesSourceidWithResponse(ctx, id, spaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return resp, nil
	case http.StatusNotFound:
		return nil, nil
	default:
		return nil, reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// CreateAgentDownloadSource creates a new agent binary download source.
func CreateAgentDownloadSource(
	ctx context.Context,
	client *Client,
	spaceID string,
	req kbapi.PostFleetAgentDownloadSourcesJSONRequestBody,
) (*kbapi.PostFleetAgentDownloadSourcesResponse, diag.Diagnostics) {
	resp, err := client.API.PostFleetAgentDownloadSourcesWithResponse(ctx, req, spaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return resp, nil
	default:
		return nil, reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// UpdateAgentDownloadSource updates an existing agent binary download source.
func UpdateAgentDownloadSource(
	ctx context.Context,
	client *Client,
	id string,
	spaceID string,
	req kbapi.PutFleetAgentDownloadSourcesSourceidJSONRequestBody,
) (*kbapi.PutFleetAgentDownloadSourcesSourceidResponse, diag.Diagnostics) {
	resp, err := client.API.PutFleetAgentDownloadSourcesSourceidWithResponse(ctx, id, req, spaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return resp, nil
	default:
		return nil, reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// DeleteAgentDownloadSource deletes an existing agent binary download source.
func DeleteAgentDownloadSource(ctx context.Context, client *Client, id string, spaceID string) diag.Diagnostics {
	resp, err := client.API.DeleteFleetAgentDownloadSourcesSourceidWithResponse(ctx, id, spaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return nil
	case http.StatusNotFound:
		return nil
	default:
		return reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// ListAgentDownloadSources reads all agent binary download sources from the API.
func ListAgentDownloadSources(ctx context.Context, client *Client, spaceID string) (*kbapi.GetFleetAgentDownloadSourcesResponse, diag.Diagnostics) {
	resp, err := client.API.GetFleetAgentDownloadSourcesWithResponse(ctx, spaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return resp, nil
	default:
		return nil, reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// GetPackage reads a specific package from the API.
func GetPackage(ctx context.Context, client *Client, name, version, spaceID string) (*kbapi.PackageInfo, diag.Diagnostics) {
	params := kbapi.GetFleetEpmPackagesPkgnamePkgversionParams{}

	resp, err := client.API.GetFleetEpmPackagesPkgnamePkgversionWithResponse(ctx, name, version, &params, spaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return &resp.JSON200.Item, nil
	case http.StatusNotFound:
		return nil, nil
	default:
		return nil, reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// InstallPackageOptions holds the options for installing a package.
type InstallPackageOptions struct {
	SpaceID                   string
	Force                     bool
	Prerelease                bool
	IgnoreMappingUpdateErrors *bool
	SkipDataStreamRollover    *bool
	IgnoreConstraints         bool
}

// InstallPackage installs a package.
func InstallPackage(ctx context.Context, client *Client, name, version string, opts InstallPackageOptions) diag.Diagnostics {
	params := kbapi.PostFleetEpmPackagesPkgnamePkgversionParams{
		Prerelease:                &opts.Prerelease,
		IgnoreMappingUpdateErrors: opts.IgnoreMappingUpdateErrors,
		SkipDataStreamRollover:    opts.SkipDataStreamRollover,
	}
	body := kbapi.PostFleetEpmPackagesPkgnamePkgversionJSONRequestBody{
		Force:             &opts.Force,
		IgnoreConstraints: &opts.IgnoreConstraints,
	}

	resp, err := client.API.PostFleetEpmPackagesPkgnamePkgversionWithResponse(ctx, name, version, &params, body, spaceAwarePathRequestEditor(opts.SpaceID))
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return nil
	default:
		return reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// Uninstall uninstalls a package.
func Uninstall(ctx context.Context, client *Client, name, version string, spaceID string, _ bool) diag.Diagnostics {
	resp, err := client.API.DeleteFleetEpmPackagesPkgnamePkgversionWithResponse(ctx, name, version, nil, spaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return nil
	case http.StatusBadRequest:
		msg := resp.JSON400.Message
		if msg == fmt.Sprintf("%s is not installed", name) {
			return nil
		}
		return reportUnknownError(resp.StatusCode(), resp.Body)
	case http.StatusNotFound:
		return nil
	default:
		return reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// GetPackages returns information about the latest packages known to Fleet.
// If spaceID is non-empty and not "default", the request will be scoped to that Kibana space.
func GetPackages(ctx context.Context, client *Client, prerelease bool, spaceID string) ([]kbapi.PackageListItem, diag.Diagnostics) {
	params := kbapi.GetFleetEpmPackagesParams{
		Prerelease: &prerelease,
	}

	resp, err := client.API.GetFleetEpmPackagesWithResponse(ctx, &params, spaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return resp.JSON200.Items, nil
	case http.StatusBadRequest:
		// Older Kibana versions (pre-8.7) do not recognise the prerelease query
		// parameter and return 400 with "definition for this key is missing".
		// Retry without the parameter so we remain compatible.
		if strings.Contains(string(resp.Body), "prerelease") {
			retryParams := kbapi.GetFleetEpmPackagesParams{}
			retryResp, retryErr := client.API.GetFleetEpmPackagesWithResponse(ctx, &retryParams, spaceAwarePathRequestEditor(spaceID))
			if retryErr != nil {
				return nil, diagutil.FrameworkDiagFromError(retryErr)
			}
			if retryResp.StatusCode() == http.StatusOK {
				return retryResp.JSON200.Items, nil
			}
			return nil, reportUnknownError(retryResp.StatusCode(), retryResp.Body)
		}
		return nil, reportUnknownError(resp.StatusCode(), resp.Body)
	default:
		return nil, reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// UploadPackageOptions holds the options for uploading a custom integration package.
type UploadPackageOptions struct {
	// PackagePath is the path to the package archive to upload (.zip or .tar.gz/.tgz).
	PackagePath string
	// ContentType is the MIME type of the package file (e.g. "application/zip").
	ContentType string
	// IgnoreMappingUpdateErrors suppresses mapping update errors during install.
	IgnoreMappingUpdateErrors bool
	// SkipDataStreamRollover skips data stream rollover during install.
	SkipDataStreamRollover bool
	// SpaceID scopes the request to a specific Kibana space.
	SpaceID string
}

// UploadPackageResult holds the result of uploading a custom integration package.
type UploadPackageResult struct {
	// PackageName is the name of the uploaded package as returned by Fleet.
	PackageName string
	// PackageVersion is the installed version resolved from the package list.
	PackageVersion string
	// AlreadyInstalled is true when Fleet rejected the upload because a package
	// with the same name is already installed. The caller should uninstall the
	// existing package and retry.
	AlreadyInstalled bool
}

// UploadPackage uploads a custom integration package to Fleet and returns the
// resolved package name and installed version. It opens the file at
// opts.PackagePath, posts it to the Fleet EPM packages endpoint, extracts the
// package name from the response, and then queries the package list to resolve
// the installed version.
func UploadPackage(ctx context.Context, client *Client, opts UploadPackageOptions) (*UploadPackageResult, diag.Diagnostics) {
	f, err := os.Open(opts.PackagePath)
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(fmt.Errorf("opening package file %q: %w", opts.PackagePath, err))
	}
	defer f.Close()

	params := kbapi.PostFleetEpmPackagesParams{
		IgnoreMappingUpdateErrors: &opts.IgnoreMappingUpdateErrors,
		SkipDataStreamRollover:    &opts.SkipDataStreamRollover,
	}

	resp, err := client.API.PostFleetEpmPackagesWithBodyWithResponse(ctx, &params, opts.ContentType, f, spaceAwarePathRequestEditor(opts.SpaceID))
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK, http.StatusCreated:
		// intentional fall-through
	default:
		// Older Kibana (8.0.x) rejects re-uploading a same-name package that is
		// already installed. Signal this condition so the caller can uninstall and
		// retry rather than treating it as a hard failure.
		if strings.Contains(string(resp.Body), "already installed") {
			return &UploadPackageResult{AlreadyInstalled: true}, nil
		}
		return nil, reportUnknownError(resp.StatusCode(), resp.Body)
	}

	// The response body does not have a typed JSON200 field; unmarshal manually.
	// The field that carries the package name and version changed across Kibana versions:
	//   - newer Kibana (8.8+): _meta.name / _meta.version
	//   - older Kibana (8.0–8.7): items[0].name / items[0].version
	//   - oldest Kibana (7.x): response[0].name / response[0].version
	// Try all three paths; if none yields a name, fall back to parsing the
	// zip manifest directly (version-independent but zip-only).
	var uploadResp struct {
		Meta struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		} `json:"_meta"`
		Items []struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		} `json:"items"`
		Response []struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		} `json:"response"`
	}
	// Best-effort unmarshal; an error here is non-fatal — we fall through to
	// the zip-manifest fallback below.
	_ = json.Unmarshal(resp.Body, &uploadResp)

	packageName := uploadResp.Meta.Name
	packageVersion := uploadResp.Meta.Version
	if packageName == "" && len(uploadResp.Items) > 0 {
		packageName = uploadResp.Items[0].Name
		packageVersion = uploadResp.Items[0].Version
	}
	if packageName == "" && len(uploadResp.Response) > 0 {
		packageName = uploadResp.Response[0].Name
		packageVersion = uploadResp.Response[0].Version
	}
	if packageName == "" {
		// Last resort: parse the name from the zip manifest. This is reliable
		// across all Kibana versions but only works for zip archives.
		packageName, err = parsePackageNameFromZip(opts.PackagePath)
		if err != nil {
			return nil, diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Invalid upload response",
					"Fleet did not return a package name and the zip manifest could not be parsed: "+err.Error(),
				),
			}
		}
	}

	// Always resolve the authoritative installed version via the package list so
	// that the stored version matches what subsequent Read calls will observe.
	// The version from the upload response body (_meta.version / items[0].version /
	// response[0].version) is captured above but intentionally not used as the
	// final source of truth because the package list is the canonical read path.
	_ = packageVersion // captured for reference; GetPackages is authoritative

	// Resolve the installed version by querying the package list and filtering by name.
	packages, diags := GetPackages(ctx, client, true, opts.SpaceID)
	if diags.HasError() {
		return nil, diags
	}

	for _, pkg := range packages {
		if pkg.Name == packageName {
			return &UploadPackageResult{
				PackageName:    packageName,
				PackageVersion: pkg.Version,
			}, nil
		}
	}

	return nil, diag.Diagnostics{
		diag.NewErrorDiagnostic(
			"Package not found after upload",
			fmt.Sprintf("Package %q was uploaded but could not be found in the package list", packageName),
		),
	}
}

// parsePackageNameFromZip opens a zip archive at path, finds the top-level
// manifest.yml, and extracts the package name field. It is used as a fallback
// when the Fleet upload API response does not include _meta.name (older Kibana).
func parsePackageNameFromZip(path string) (string, error) {
	r, err := zip.OpenReader(path)
	if err != nil {
		return "", fmt.Errorf("opening zip %q: %w", path, err)
	}
	defer r.Close()

	nameRe := regexp.MustCompile(`(?m)^name:\s*(\S+)`)
	for _, f := range r.File {
		if !strings.HasSuffix(f.Name, "/manifest.yml") && f.Name != "manifest.yml" {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return "", fmt.Errorf("opening manifest.yml in zip: %w", err)
		}
		content, readErr := io.ReadAll(rc)
		rc.Close()
		if readErr != nil {
			return "", fmt.Errorf("reading manifest.yml: %w", readErr)
		}
		matches := nameRe.FindSubmatch(content)
		if len(matches) >= 2 {
			return string(matches[1]), nil
		}
	}
	return "", fmt.Errorf("manifest.yml with name field not found in zip")
}
