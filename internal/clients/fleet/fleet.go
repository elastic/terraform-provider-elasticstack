package fleet

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
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
	return func(ctx context.Context, req *http.Request) error {
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
		SysMonitoring: utils.Pointer(sysMonitoring),
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
		Format: utils.Pointer(kbapi.GetFleetPackagePoliciesPackagepolicyidParamsFormatSimplified),
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

// CreatePackagePolicy creates a new package policy.
func CreatePackagePolicy(ctx context.Context, client *Client, spaceID string, req kbapi.PackagePolicyRequest) (*kbapi.PackagePolicy, diag.Diagnostics) {
	params := kbapi.PostFleetPackagePoliciesParams{
		Format: utils.Pointer(kbapi.PostFleetPackagePoliciesParamsFormatSimplified),
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

// UpdatePackagePolicy updates an existing package policy.
func UpdatePackagePolicy(ctx context.Context, client *Client, id string, spaceID string, req kbapi.PackagePolicyRequest) (*kbapi.PackagePolicy, diag.Diagnostics) {
	params := kbapi.PutFleetPackagePoliciesPackagepolicyidParams{
		Format: utils.Pointer(kbapi.Simplified),
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
func Uninstall(ctx context.Context, client *Client, name, version string, spaceID string, force bool) diag.Diagnostics {
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
