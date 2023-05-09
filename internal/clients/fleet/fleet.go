package fleet

import (
	"context"
	"fmt"
	"net/http"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet/fleetapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

// AllEnrollmentTokens reads all enrollment tokens from the API.
func AllEnrollmentTokens(ctx context.Context, client *Client) ([]fleetapi.EnrollmentApiKey, diag.Diagnostics) {
	resp, err := client.API.GetEnrollmentApiKeysWithResponse(ctx)
	if err != nil {
		return nil, diag.FromErr(err)
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
		return nil, diag.FromErr(err)
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
func CreateAgentPolicy(ctx context.Context, client *Client, req fleetapi.AgentPolicyCreateRequest) (*fleetapi.AgentPolicy, diag.Diagnostics) {
	resp, err := client.API.CreateAgentPolicyWithResponse(ctx, req)
	if err != nil {
		return nil, diag.FromErr(err)
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
		return nil, diag.FromErr(err)
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
		return diag.FromErr(err)
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

func reportUnknownError(statusCode int, body []byte) diag.Diagnostics {
	return diag.Diagnostics{
		diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Unexpected status code from server: got HTTP %d", statusCode),
			Detail:   string(body),
		},
	}
}
