package fleet

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

var (
	ErrPackageNotFound = errors.New("package not found")
)

// GetEnrollmentTokens reads all enrollment tokens from the API.
func GetEnrollmentTokens(ctx context.Context, client *Client) ([]kbapi.EnrollmentApiKey, diag.Diagnostics) {
	resp, err := client.API.GetFleetEnrollmentApiKeysWithResponse(ctx, nil)
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
		Kuery: utils.Pointer("policy_id:" + policyID),
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

// GetAgentPolicy reads a specific agent policy from the API.
func GetAgentPolicy(ctx context.Context, client *Client, id string) (*kbapi.AgentPolicy, diag.Diagnostics) {
	resp, err := client.API.GetFleetAgentPoliciesAgentpolicyidWithResponse(ctx, id, nil)
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

// GetAgentPolicyInSpace reads a specific agent policy from the API within a specific Kibana space.
// This is necessary for space-aware policies that are created with space_ids.
func GetAgentPolicyInSpace(ctx context.Context, client *Client, id string, spaceID string) (*kbapi.AgentPolicy, diag.Diagnostics) {
	// Construct the space-aware path
	// For default space: /api/fleet/agent_policies/{id}
	// For custom space: /s/{space_id}/api/fleet/agent_policies/{id}
	path := fmt.Sprintf("/api/fleet/agent_policies/%s", id)
	if spaceID != "" && spaceID != "default" {
		path = fmt.Sprintf("/s/%s/api/fleet/agent_policies/%s", spaceID, id)
	}

	// Make the request using the underlying HTTP client
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
			Item kbapi.AgentPolicy `json:"item"`
		}
		if err := json.NewDecoder(httpResp.Body).Decode(&result); err != nil {
			return nil, diagutil.FrameworkDiagFromError(err)
		}
		return &result.Item, nil
	case http.StatusNotFound:
		return nil, nil
	default:
		bodyBytes, _ := io.ReadAll(httpResp.Body)
		return nil, reportUnknownError(httpResp.StatusCode, bodyBytes)
	}
}

// CreateAgentPolicy creates a new agent policy.
func CreateAgentPolicy(ctx context.Context, client *Client, req kbapi.PostFleetAgentPoliciesJSONRequestBody, sysMonitoring bool) (*kbapi.AgentPolicy, diag.Diagnostics) {
	params := kbapi.PostFleetAgentPoliciesParams{
		SysMonitoring: utils.Pointer(sysMonitoring),
	}

	resp, err := client.API.PostFleetAgentPoliciesWithResponse(ctx, &params, req)
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
func UpdateAgentPolicy(ctx context.Context, client *Client, id string, req kbapi.PutFleetAgentPoliciesAgentpolicyidJSONRequestBody) (*kbapi.AgentPolicy, diag.Diagnostics) {
	resp, err := client.API.PutFleetAgentPoliciesAgentpolicyidWithResponse(ctx, id, nil, req)
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

// UpdateAgentPolicyInSpace updates an existing agent policy within a specific Kibana space.
// This is necessary for space-aware policies that are created with space_ids.
func UpdateAgentPolicyInSpace(ctx context.Context, client *Client, id string, spaceID string, reqBody kbapi.PutFleetAgentPoliciesAgentpolicyidJSONRequestBody) (*kbapi.AgentPolicy, diag.Diagnostics) {
	// Construct the space-aware path
	path := fmt.Sprintf("/api/fleet/agent_policies/%s", id)
	if spaceID != "" && spaceID != "default" {
		path = fmt.Sprintf("/s/%s/api/fleet/agent_policies/%s", spaceID, id)
	}

	// Marshal the request body
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	// Make the request using the underlying HTTP client
	req, err := http.NewRequestWithContext(ctx, "PUT", client.URL+path, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}
	req.Header.Set("Content-Type", "application/json")

	httpResp, err := client.HTTP.Do(req)
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}
	defer httpResp.Body.Close()

	switch httpResp.StatusCode {
	case http.StatusOK:
		var result struct {
			Item kbapi.AgentPolicy `json:"item"`
		}
		if err := json.NewDecoder(httpResp.Body).Decode(&result); err != nil {
			return nil, diagutil.FrameworkDiagFromError(err)
		}
		return &result.Item, nil
	default:
		bodyBytes, _ := io.ReadAll(httpResp.Body)
		return nil, reportUnknownError(httpResp.StatusCode, bodyBytes)
	}
}

// DeleteAgentPolicy deletes an existing agent policy.
func DeleteAgentPolicy(ctx context.Context, client *Client, id string) diag.Diagnostics {
	body := kbapi.PostFleetAgentPoliciesDeleteJSONRequestBody{
		AgentPolicyId: id,
	}

	resp, err := client.API.PostFleetAgentPoliciesDeleteWithResponse(ctx, body)
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

// DeleteAgentPolicyInSpace deletes an existing agent policy within a specific Kibana space.
// This is necessary for space-aware policies that are created with space_ids.
func DeleteAgentPolicyInSpace(ctx context.Context, client *Client, id string, spaceID string) diag.Diagnostics {
	// Construct the space-aware path
	path := "/api/fleet/agent_policies/delete"
	if spaceID != "" && spaceID != "default" {
		path = fmt.Sprintf("/s/%s/api/fleet/agent_policies/delete", spaceID)
	}

	// Create request body
	reqBody := kbapi.PostFleetAgentPoliciesDeleteJSONRequestBody{
		AgentPolicyId: id,
	}

	// Marshal the request body
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	// Make the request using the underlying HTTP client
	req, err := http.NewRequestWithContext(ctx, "POST", client.URL+path, bytes.NewReader(bodyBytes))
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	req.Header.Set("Content-Type", "application/json")

	httpResp, err := client.HTTP.Do(req)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	defer httpResp.Body.Close()

	switch httpResp.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusNotFound:
		return nil
	default:
		bodyBytes, _ := io.ReadAll(httpResp.Body)
		return reportUnknownError(httpResp.StatusCode, bodyBytes)
	}
}

// GetOutput reads a specific output from the API.
func GetOutput(ctx context.Context, client *Client, id string) (*kbapi.OutputUnion, diag.Diagnostics) {
	resp, err := client.API.GetFleetOutputsOutputidWithResponse(ctx, id)
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
func CreateOutput(ctx context.Context, client *Client, req kbapi.NewOutputUnion) (*kbapi.OutputUnion, diag.Diagnostics) {
	resp, err := client.API.PostFleetOutputsWithResponse(ctx, req)
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
func UpdateOutput(ctx context.Context, client *Client, id string, req kbapi.UpdateOutputUnion) (*kbapi.OutputUnion, diag.Diagnostics) {
	resp, err := client.API.PutFleetOutputsOutputidWithResponse(ctx, id, req)
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
func DeleteOutput(ctx context.Context, client *Client, id string) diag.Diagnostics {
	resp, err := client.API.DeleteFleetOutputsOutputidWithResponse(ctx, id)
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
func GetFleetServerHost(ctx context.Context, client *Client, id string) (*kbapi.ServerHost, diag.Diagnostics) {
	resp, err := client.API.GetFleetFleetServerHostsItemidWithResponse(ctx, id)
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
func CreateFleetServerHost(ctx context.Context, client *Client, req kbapi.PostFleetFleetServerHostsJSONRequestBody) (*kbapi.ServerHost, diag.Diagnostics) {
	resp, err := client.API.PostFleetFleetServerHostsWithResponse(ctx, req)
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
func UpdateFleetServerHost(ctx context.Context, client *Client, id string, req kbapi.PutFleetFleetServerHostsItemidJSONRequestBody) (*kbapi.ServerHost, diag.Diagnostics) {
	resp, err := client.API.PutFleetFleetServerHostsItemidWithResponse(ctx, id, req)
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
func DeleteFleetServerHost(ctx context.Context, client *Client, id string) diag.Diagnostics {
	resp, err := client.API.DeleteFleetFleetServerHostsItemidWithResponse(ctx, id)
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
func GetPackagePolicy(ctx context.Context, client *Client, id string) (*kbapi.PackagePolicy, diag.Diagnostics) {
	params := kbapi.GetFleetPackagePoliciesPackagepolicyidParams{
		Format: utils.Pointer(kbapi.GetFleetPackagePoliciesPackagepolicyidParamsFormatSimplified),
	}

	resp, err := client.API.GetFleetPackagePoliciesPackagepolicyidWithResponse(ctx, id, &params)
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

// GetPackagePolicyInSpace reads a specific package policy from the API within a specific Kibana space.
// This is necessary for space-aware policies that are created with space_ids.
func GetPackagePolicyInSpace(ctx context.Context, client *Client, id string, spaceID string) (*kbapi.PackagePolicy, diag.Diagnostics) {
	// Construct the space-aware path
	path := fmt.Sprintf("/api/fleet/package_policies/%s?format=simplified", id)
	if spaceID != "" && spaceID != "default" {
		path = fmt.Sprintf("/s/%s/api/fleet/package_policies/%s?format=simplified", spaceID, id)
	}

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
			Item kbapi.PackagePolicy `json:"item"`
		}
		if err := json.NewDecoder(httpResp.Body).Decode(&result); err != nil {
			return nil, diagutil.FrameworkDiagFromError(err)
		}
		return &result.Item, nil
	case http.StatusNotFound:
		return nil, nil
	default:
		bodyBytes, _ := io.ReadAll(httpResp.Body)
		return nil, reportUnknownError(httpResp.StatusCode, bodyBytes)
	}
}

// CreatePackagePolicy creates a new package policy.
func CreatePackagePolicy(ctx context.Context, client *Client, req kbapi.PackagePolicyRequest) (*kbapi.PackagePolicy, diag.Diagnostics) {
	params := kbapi.PostFleetPackagePoliciesParams{
		Format: utils.Pointer(kbapi.PostFleetPackagePoliciesParamsFormatSimplified),
	}

	resp, err := client.API.PostFleetPackagePoliciesWithResponse(ctx, &params, req)
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

// CreatePackagePolicyInSpace creates a new package policy within a specific Kibana space.
// This is necessary when the referenced agent policy exists in a specific space.
func CreatePackagePolicyInSpace(ctx context.Context, client *Client, spaceID string, reqBody kbapi.PackagePolicyRequest) (*kbapi.PackagePolicy, diag.Diagnostics) {
	// Construct the space-aware path
	path := "/api/fleet/package_policies?format=simplified"
	if spaceID != "" && spaceID != "default" {
		path = fmt.Sprintf("/s/%s/api/fleet/package_policies?format=simplified", spaceID)
	}

	// Marshal the request body
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", client.URL+path, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}
	req.Header.Set("Content-Type", "application/json")

	httpResp, err := client.HTTP.Do(req)
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}
	defer httpResp.Body.Close()

	switch httpResp.StatusCode {
	case http.StatusOK:
		var result struct {
			Item kbapi.PackagePolicy `json:"item"`
		}
		if err := json.NewDecoder(httpResp.Body).Decode(&result); err != nil {
			return nil, diagutil.FrameworkDiagFromError(err)
		}
		return &result.Item, nil
	default:
		bodyBytes, _ := io.ReadAll(httpResp.Body)
		return nil, reportUnknownError(httpResp.StatusCode, bodyBytes)
	}
}

// UpdatePackagePolicy updates an existing package policy.
func UpdatePackagePolicy(ctx context.Context, client *Client, id string, req kbapi.PackagePolicyRequest) (*kbapi.PackagePolicy, diag.Diagnostics) {
	params := kbapi.PutFleetPackagePoliciesPackagepolicyidParams{
		Format: utils.Pointer(kbapi.PutFleetPackagePoliciesPackagepolicyidParamsFormatSimplified),
	}

	resp, err := client.API.PutFleetPackagePoliciesPackagepolicyidWithResponse(ctx, id, &params, req)
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

// UpdatePackagePolicyInSpace updates an existing package policy within a specific Kibana space.
// This is necessary for space-aware policies that are created with space_ids.
func UpdatePackagePolicyInSpace(ctx context.Context, client *Client, id string, spaceID string, reqBody kbapi.PackagePolicyRequest) (*kbapi.PackagePolicy, diag.Diagnostics) {
	// Construct the space-aware path
	path := fmt.Sprintf("/api/fleet/package_policies/%s?format=simplified", id)
	if spaceID != "" && spaceID != "default" {
		path = fmt.Sprintf("/s/%s/api/fleet/package_policies/%s?format=simplified", spaceID, id)
	}

	// Marshal the request body
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", client.URL+path, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}
	req.Header.Set("Content-Type", "application/json")

	httpResp, err := client.HTTP.Do(req)
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}
	defer httpResp.Body.Close()

	switch httpResp.StatusCode {
	case http.StatusOK:
		var result struct {
			Item kbapi.PackagePolicy `json:"item"`
		}
		if err := json.NewDecoder(httpResp.Body).Decode(&result); err != nil {
			return nil, diagutil.FrameworkDiagFromError(err)
		}
		return &result.Item, nil
	default:
		bodyBytes, _ := io.ReadAll(httpResp.Body)
		return nil, reportUnknownError(httpResp.StatusCode, bodyBytes)
	}
}

// DeletePackagePolicy deletes an existing package policy.
func DeletePackagePolicy(ctx context.Context, client *Client, id string, force bool) diag.Diagnostics {
	params := kbapi.DeleteFleetPackagePoliciesPackagepolicyidParams{
		Force: &force,
	}

	resp, err := client.API.DeleteFleetPackagePoliciesPackagepolicyidWithResponse(ctx, id, &params)
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

// DeletePackagePolicyInSpace deletes an existing package policy within a specific Kibana space.
// This is necessary for space-aware policies that are created with space_ids.
func DeletePackagePolicyInSpace(ctx context.Context, client *Client, id string, spaceID string, force bool) diag.Diagnostics {
	// Construct the space-aware path
	path := fmt.Sprintf("/api/fleet/package_policies/%s?force=%t", id, force)
	if spaceID != "" && spaceID != "default" {
		path = fmt.Sprintf("/s/%s/api/fleet/package_policies/%s?force=%t", spaceID, id, force)
	}

	req, err := http.NewRequestWithContext(ctx, "DELETE", client.URL+path, nil)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	httpResp, err := client.HTTP.Do(req)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	defer httpResp.Body.Close()

	switch httpResp.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusNotFound:
		return nil
	default:
		bodyBytes, _ := io.ReadAll(httpResp.Body)
		return reportUnknownError(httpResp.StatusCode, bodyBytes)
	}
}

// GetPackage reads a specific package from the API.
func GetPackage(ctx context.Context, client *Client, name, version string) (*kbapi.PackageInfo, diag.Diagnostics) {
	params := kbapi.GetFleetEpmPackagesPkgnamePkgversionParams{}

	resp, err := client.API.GetFleetEpmPackagesPkgnamePkgversionWithResponse(ctx, name, version, &params)
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return &resp.JSON200.Item, nil
	case http.StatusNotFound:
		return nil, diagutil.FrameworkDiagFromError(ErrPackageNotFound)
	default:
		return nil, reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// InstallPackage installs a package.
func InstallPackage(ctx context.Context, client *Client, name, version string, force bool) diag.Diagnostics {
	params := kbapi.PostFleetEpmPackagesPkgnamePkgversionParams{}
	body := kbapi.PostFleetEpmPackagesPkgnamePkgversionJSONRequestBody{
		Force: &force,
	}

	resp, err := client.API.PostFleetEpmPackagesPkgnamePkgversionWithResponse(ctx, name, version, &params, body)
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
func Uninstall(ctx context.Context, client *Client, name, version string, force bool) diag.Diagnostics {
	resp, err := client.API.DeleteFleetEpmPackagesPkgnamePkgversionWithResponse(ctx, name, version, nil)
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
		} else {
			return reportUnknownError(resp.StatusCode(), resp.Body)
		}
	case http.StatusNotFound:
		return nil
	default:
		return reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// GetPackages returns information about the latest packages known to Fleet.
func GetPackages(ctx context.Context, client *Client, prerelease bool) ([]kbapi.PackageListItem, diag.Diagnostics) {
	params := kbapi.GetFleetEpmPackagesParams{
		Prerelease: &prerelease,
	}

	resp, err := client.API.GetFleetEpmPackagesWithResponse(ctx, &params)
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
