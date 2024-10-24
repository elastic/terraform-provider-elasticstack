package fleet

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	fleetapi "github.com/elastic/terraform-provider-elasticstack/generated/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

var (
	ErrPackageNotFound = errors.New("package not found")
)

// GetEnrollmentTokens reads all enrollment tokens from the API.
func GetEnrollmentTokens(ctx context.Context, client *Client) ([]fleetapi.EnrollmentApiKey, diag.Diagnostics) {
	resp, err := client.API.GetEnrollmentApiKeysWithResponse(ctx, nil)
	if err != nil {
		return nil, utils.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return resp.JSON200.Items, nil
	default:
		return nil, reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// GetEnrollmentTokensByPolicy Get enrollment tokens by given policy ID
func GetEnrollmentTokensByPolicy(ctx context.Context, client *Client, policyID string) ([]fleetapi.EnrollmentApiKey, diag.Diagnostics) {
	params := fleetapi.GetEnrollmentApiKeysParams{
		Kuery: utils.Pointer("policy_id:" + policyID),
	}

	resp, err := client.API.GetEnrollmentApiKeysWithResponse(ctx, &params)
	if err != nil {
		return nil, utils.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return resp.JSON200.Items, nil
	default:
		return nil, reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// GetAgentPolicy reads a specific agent policy from the API.
func GetAgentPolicy(ctx context.Context, client *Client, id string) (*fleetapi.AgentPolicy, diag.Diagnostics) {
	resp, err := client.API.GetAgentPolicyWithResponse(ctx, id, nil)
	if err != nil {
		return nil, utils.FrameworkDiagFromError(err)
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
func CreateAgentPolicy(ctx context.Context, client *Client, req fleetapi.CreateAgentPolicyJSONRequestBody, sysMonitoring bool) (*fleetapi.AgentPolicy, diag.Diagnostics) {
	params := fleetapi.CreateAgentPolicyParams{
		SysMonitoring: utils.Pointer(sysMonitoring),
	}

	resp, err := client.API.CreateAgentPolicyWithResponse(ctx, &params, req)
	if err != nil {
		return nil, utils.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return &resp.JSON200.Item, nil
	default:
		return nil, reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// UpdateAgentPolicy updates an existing agent policy.
func UpdateAgentPolicy(ctx context.Context, client *Client, id string, req fleetapi.UpdateAgentPolicyJSONRequestBody) (*fleetapi.AgentPolicy, diag.Diagnostics) {
	resp, err := client.API.UpdateAgentPolicyWithResponse(ctx, id, nil, req)
	if err != nil {
		return nil, utils.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return &resp.JSON200.Item, nil
	default:
		return nil, reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// DeleteAgentPolicy deletes an existing agent policy
func DeleteAgentPolicy(ctx context.Context, client *Client, id string) diag.Diagnostics {
	body := fleetapi.DeleteAgentPolicyJSONRequestBody{
		AgentPolicyId: id,
	}

	resp, err := client.API.DeleteAgentPolicyWithResponse(ctx, body)
	if err != nil {
		return utils.FrameworkDiagFromError(err)
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
func GetOutput(ctx context.Context, client *Client, id string) (*fleetapi.OutputUnion, diag.Diagnostics) {
	resp, err := client.API.GetOutputWithResponse(ctx, id)
	if err != nil {
		return nil, utils.FrameworkDiagFromError(err)
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
func CreateOutput(ctx context.Context, client *Client, req fleetapi.NewOutputUnion) (*fleetapi.OutputUnion, diag.Diagnostics) {
	resp, err := client.API.CreateOutputWithResponse(ctx, req)
	if err != nil {
		return nil, utils.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return &resp.JSON200.Item, nil
	default:
		return nil, reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// UpdateOutput updates an existing output.
func UpdateOutput(ctx context.Context, client *Client, id string, req fleetapi.UpdateOutputUnion) (*fleetapi.OutputUnion, diag.Diagnostics) {
	resp, err := client.API.UpdateOutputWithResponse(ctx, id, req)
	if err != nil {
		return nil, utils.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return &resp.JSON200.Item, nil
	default:
		return nil, reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// DeleteOutput deletes an existing output
func DeleteOutput(ctx context.Context, client *Client, id string) diag.Diagnostics {
	resp, err := client.API.DeleteOutputWithResponse(ctx, id)
	if err != nil {
		return utils.FrameworkDiagFromError(err)
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
func GetFleetServerHost(ctx context.Context, client *Client, id string) (*fleetapi.ServerHost, diag.Diagnostics) {
	resp, err := client.API.GetFleetServerHostWithResponse(ctx, id)
	if err != nil {
		return nil, utils.FrameworkDiagFromError(err)
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
func CreateFleetServerHost(ctx context.Context, client *Client, req fleetapi.CreateFleetServerHostJSONRequestBody) (*fleetapi.ServerHost, diag.Diagnostics) {
	resp, err := client.API.CreateFleetServerHostWithResponse(ctx, req)
	if err != nil {
		return nil, utils.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return &resp.JSON200.Item, nil
	default:
		return nil, reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// UpdateFleetServerHost updates an existing fleet server host.
func UpdateFleetServerHost(ctx context.Context, client *Client, id string, req fleetapi.UpdateFleetServerHostJSONRequestBody) (*fleetapi.ServerHost, diag.Diagnostics) {
	resp, err := client.API.UpdateFleetServerHostWithResponse(ctx, id, req)
	if err != nil {
		return nil, utils.FrameworkDiagFromError(err)
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
	resp, err := client.API.DeleteFleetServerHostWithResponse(ctx, id)
	if err != nil {
		return utils.FrameworkDiagFromError(err)
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
func GetPackagePolicy(ctx context.Context, client *Client, id string) (*fleetapi.PackagePolicy, diag.Diagnostics) {
	params := fleetapi.GetPackagePolicyParams{
		Format: utils.Pointer(fleetapi.GetPackagePolicyParamsFormatSimplified),
	}

	resp, err := client.API.GetPackagePolicyWithResponse(ctx, id, &params)
	if err != nil {
		return nil, utils.FrameworkDiagFromError(err)
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
func CreatePackagePolicy(ctx context.Context, client *Client, req fleetapi.CreatePackagePolicyJSONRequestBody) (*fleetapi.PackagePolicy, diag.Diagnostics) {
	params := fleetapi.CreatePackagePolicyParams{
		Format: utils.Pointer(fleetapi.CreatePackagePolicyParamsFormatSimplified),
	}

	resp, err := client.API.CreatePackagePolicyWithResponse(ctx, &params, req)
	if err != nil {
		return nil, utils.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return &resp.JSON200.Item, nil
	default:
		return nil, reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// UpdatePackagePolicy updates an existing package policy.
func UpdatePackagePolicy(ctx context.Context, client *Client, id string, req fleetapi.UpdatePackagePolicyJSONRequestBody) (*fleetapi.PackagePolicy, diag.Diagnostics) {
	params := fleetapi.UpdatePackagePolicyParams{
		Format: utils.Pointer(fleetapi.Simplified),
	}

	resp, err := client.API.UpdatePackagePolicyWithResponse(ctx, id, &params, req)
	if err != nil {
		return nil, utils.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return &resp.JSON200.Item, nil
	default:
		return nil, reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// DeletePackagePolicy deletes an existing package policy.
func DeletePackagePolicy(ctx context.Context, client *Client, id string, force bool) diag.Diagnostics {
	params := fleetapi.DeletePackagePolicyParams{
		Force: &force,
	}

	resp, err := client.API.DeletePackagePolicyWithResponse(ctx, id, &params)
	if err != nil {
		return utils.FrameworkDiagFromError(err)
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
func GetPackage(ctx context.Context, client *Client, name, version string) (*fleetapi.PackageInfo, diag.Diagnostics) {
	params := fleetapi.GetPackageParams{}

	resp, err := client.API.GetPackageWithResponse(ctx, name, version, &params)
	if err != nil {
		return nil, utils.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return &resp.JSON200.Item, nil
	case http.StatusNotFound:
		return nil, utils.FrameworkDiagFromError(ErrPackageNotFound)
	default:
		return nil, reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// InstallPackage installs a package.
func InstallPackage(ctx context.Context, client *Client, name, version string, force bool) diag.Diagnostics {
	params := fleetapi.InstallPackageParams{}
	body := fleetapi.InstallPackageJSONRequestBody{
		Force: &force,
	}

	resp, err := client.API.InstallPackageWithResponse(ctx, name, version, &params, body)
	if err != nil {
		return utils.FrameworkDiagFromError(err)
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
	body := fleetapi.DeletePackageJSONRequestBody{
		Force: force,
	}

	resp, err := client.API.DeletePackageWithResponse(ctx, name, version, nil, body)
	if err != nil {
		return utils.FrameworkDiagFromError(err)
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
func GetPackages(ctx context.Context, client *Client, prerelease bool) ([]fleetapi.PackageListItem, diag.Diagnostics) {
	params := fleetapi.ListPackagesParams{
		Prerelease: &prerelease,
	}

	resp, err := client.API.ListPackagesWithResponse(ctx, &params)
	if err != nil {
		return nil, utils.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return resp.JSON200.Items, nil
	default:
		return nil, reportUnknownError(resp.StatusCode(), resp.Body)
	}
}
