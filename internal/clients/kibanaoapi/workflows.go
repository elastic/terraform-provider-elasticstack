package kibanaoapi

import (
	"context"
	"net/http"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// addInternalOriginHeader adds the required x-elastic-internal-origin header for workflow APIs
func addInternalOriginHeader(ctx context.Context, req *http.Request) error {
	req.Header.Set("x-elastic-internal-origin", "Kibana")
	return nil
}

// GetWorkflow reads a specific workflow from the API.
func GetWorkflow(ctx context.Context, client *Client, workflowID string) (*kbapi.WorkflowDetailDto, diag.Diagnostics) {
	resp, err := client.API.GetWorkflowsIdWithResponse(ctx, workflowID, addInternalOriginHeader)
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return resp.JSON200, nil
	case http.StatusNotFound:
		return nil, nil
	default:
		return nil, reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// CreateWorkflow creates a new workflow.
func CreateWorkflow(ctx context.Context, client *Client, req kbapi.CreateWorkflowCommand) (*kbapi.WorkflowDetailDto, diag.Diagnostics) {
	resp, err := client.API.PostWorkflowsWithResponse(ctx, req, addInternalOriginHeader)
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return resp.JSON200, nil
	default:
		return nil, reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// UpdateWorkflow updates an existing workflow.
func UpdateWorkflow(ctx context.Context, client *Client, workflowID string, req kbapi.UpdateWorkflowCommand) (*kbapi.UpdatedWorkflowResponseDto, diag.Diagnostics) {
	resp, err := client.API.PutWorkflowsIdWithResponse(ctx, workflowID, req, addInternalOriginHeader)
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return resp.JSON200, nil
	default:
		return nil, reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// DeleteWorkflow deletes an existing workflow.
func DeleteWorkflow(ctx context.Context, client *Client, workflowID string) diag.Diagnostics {
	resp, err := client.API.DeleteWorkflowsIdWithResponse(ctx, workflowID, addInternalOriginHeader)
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
