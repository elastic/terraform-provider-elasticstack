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

// AllEnrollmentTokens reads all enrollment tokens from the API.
func AllEnrollmentTokens(ctx context.Context, client *Client) ([]fleetapi.EnrollmentApiKey, diag.Diagnostics) {
	resp, err := client.API.GetEnrollmentApiKeysWithResponse(ctx)
	if err != nil {
		return nil, utils.FrameworkDiagFromError(err)
	}

	if resp.StatusCode() == http.StatusOK {
		return resp.JSON200.Items, nil
	}
	return nil, reportUnknownError(resp.StatusCode(), resp.Body)
}

// GetEnrollmentTokensByPolicy Get enrollment tokens by given policy ID
func GetEnrollmentTokensByPolicy(ctx context.Context, client *Client, policyID string) ([]fleetapi.EnrollmentApiKey, diag.Diagnostics) {
	resp, err := client.API.GetEnrollmentApiKeysWithResponse(ctx, func(ctx context.Context, req *http.Request) error {
		q := req.URL.Query()
		q.Set("kuery", "policy_id:"+policyID)
		req.URL.RawQuery = q.Encode()

		return nil
	})
	if err != nil {
		return nil, utils.FrameworkDiagFromError(err)
	}

	if resp.StatusCode() == http.StatusOK {
		return resp.JSON200.Items, nil
	}
	return nil, reportUnknownError(resp.StatusCode(), resp.Body)
}

// ReadAgentPolicy reads a specific agent policy from the API.
func ReadAgentPolicy(ctx context.Context, client *Client, id string) (*fleetapi.AgentPolicy, diag.Diagnostics) {
	resp, err := client.API.AgentPolicyInfoWithResponse(ctx, id)
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
func CreateAgentPolicy(ctx context.Context, client *Client, req fleetapi.AgentPolicyCreateRequest, sysMonitoring bool) (*fleetapi.AgentPolicy, diag.Diagnostics) {
	resp, err := client.API.CreateAgentPolicyWithResponse(ctx, req, func(ctx context.Context, req *http.Request) error {
		if sysMonitoring {
			qs := req.URL.Query()
			qs.Add("sys_monitoring", "true")
			req.URL.RawQuery = qs.Encode()
		}

		return nil
	})
	if err != nil {
		return nil, utils.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return resp.JSON200.Item, nil
	default:
		return nil, reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// UpdateAgentPolicy updates an existing agent policy.
func UpdateAgentPolicy(ctx context.Context, client *Client, id string, req fleetapi.AgentPolicyUpdateRequest) (*fleetapi.AgentPolicy, diag.Diagnostics) {
	resp, err := client.API.UpdateAgentPolicyWithResponse(ctx, id, req)
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

// ReadOutput reads a specific output from the API.
func ReadOutput(ctx context.Context, client *Client, id string) (*fleetapi.OutputCreateRequest, diag.Diagnostics) {
	resp, err := client.API.GetOutputWithResponse(ctx, id)
	if err != nil {
		return nil, utils.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return resp.JSON200.Item, nil
	case http.StatusNotFound:
		return nil, nil
	default:
		return nil, reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// CreateOutput creates a new output.
func CreateOutput(ctx context.Context, client *Client, req fleetapi.PostOutputsJSONRequestBody) (*fleetapi.OutputCreateRequest, diag.Diagnostics) {
	resp, err := client.API.PostOutputsWithResponse(ctx, req)
	if err != nil {
		return nil, utils.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return resp.JSON200.Item, nil
	default:
		return nil, reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// UpdateOutput updates an existing output.
func UpdateOutput(ctx context.Context, client *Client, id string, req fleetapi.UpdateOutputJSONRequestBody) (*fleetapi.OutputUpdateRequest, diag.Diagnostics) {
	resp, err := client.API.UpdateOutputWithResponse(ctx, id, req)
	if err != nil {
		return nil, utils.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return resp.JSON200.Item, nil
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

// ReadFleetServerHost reads a specific fleet server host from the API.
func ReadFleetServerHost(ctx context.Context, client *Client, id string) (*fleetapi.FleetServerHost, diag.Diagnostics) {
	resp, err := client.API.GetOneFleetServerHostsWithResponse(ctx, id)
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
func CreateFleetServerHost(ctx context.Context, client *Client, req fleetapi.PostFleetServerHostsJSONRequestBody) (*fleetapi.FleetServerHost, diag.Diagnostics) {
	resp, err := client.API.PostFleetServerHostsWithResponse(ctx, req)
	if err != nil {
		return nil, utils.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return resp.JSON200.Item, nil
	default:
		return nil, reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// UpdateFleetServerHost updates an existing fleet server host.
func UpdateFleetServerHost(ctx context.Context, client *Client, id string, req fleetapi.UpdateFleetServerHostsJSONRequestBody) (*fleetapi.FleetServerHost, diag.Diagnostics) {
	resp, err := client.API.UpdateFleetServerHostsWithResponse(ctx, id, req)
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
	resp, err := client.API.DeleteFleetServerHostsWithResponse(ctx, id)
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

// ReadPackagePolicy reads a specific package policy from the API.
func ReadPackagePolicy(ctx context.Context, client *Client, id string) (*fleetapi.PackagePolicy, diag.Diagnostics) {
	format := fleetapi.GetPackagePolicyParamsFormatSimplified
	params := fleetapi.GetPackagePolicyParams{
		Format: &format,
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
	format := fleetapi.CreatePackagePolicyParamsFormatSimplified
	params := fleetapi.CreatePackagePolicyParams{
		Format: &format,
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
	format := fleetapi.UpdatePackagePolicyParamsFormatSimplified
	params := fleetapi.UpdatePackagePolicyParams{
		Format: &format,
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
	params := fleetapi.DeletePackagePolicyParams{Force: &force}
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

// ReadPackage reads a specific package from the API.
func ReadPackage(ctx context.Context, client *Client, name, version string) (*fleetapi.GetPackageItem, diag.Diagnostics) {
	params := fleetapi.GetPackageParams{}

	resp, err := client.API.GetPackageWithResponse(ctx, name, version, &params)
	if err != nil {
		return nil, utils.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return resp.JSON200.Item, nil
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
		Force:             &force,
		IgnoreConstraints: nil,
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
	params := fleetapi.DeletePackageParams{}
	body := fleetapi.DeletePackageJSONRequestBody{
		Force: &force,
	}

	resp, err := client.API.DeletePackageWithResponse(ctx, name, version, &params, body)
	if err != nil {
		return utils.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return nil
	case http.StatusBadRequest:
		msg := resp.JSON400.Message
		if msg != nil && *msg == fmt.Sprintf("%s is not installed", name) {
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

// AllPackages returns information about the latest packages known to Fleet.
func AllPackages(ctx context.Context, client *Client, prerelease bool) ([]fleetapi.SearchResult, diag.Diagnostics) {
	params := fleetapi.ListAllPackagesParams{
		Prerelease: &prerelease,
	}

	resp, err := client.API.ListAllPackagesWithResponse(ctx, &params)
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

func reportUnknownError(statusCode int, body []byte) diag.Diagnostics {
	return diag.Diagnostics{
		diag.NewErrorDiagnostic(
			fmt.Sprintf("Unexpected status code from server: got HTTP %d", statusCode),
			string(body),
		),
	}
}
