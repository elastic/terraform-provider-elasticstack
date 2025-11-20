package kibana_oapi

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbstreams"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// newStreamsClient constructs a Streams-only API client on top of the shared
// Kibana HTTP client. This keeps Streams experimentation isolated from the
// main kbapi client.
func newStreamsClient(c *Client) (*kbstreams.ClientWithResponses, diag.Diagnostics) {
	var diags diag.Diagnostics

	endpoint := c.URL
	if !strings.HasSuffix(endpoint, "/") {
		endpoint += "/"
	}

	streamsClient, err := kbstreams.NewClientWithResponses(endpoint, kbstreams.WithHTTPClient(c.HTTP))
	if err != nil {
		diags.AddError("Failed to create Kibana Streams client", err.Error())
		return nil, diags
	}

	return streamsClient, diags
}

// GetStreamJSON reads a single stream definition (GET /api/streams/{name}).
// Returns nil, nil on 404.
func GetStreamJSON(ctx context.Context, client *Client, name string) ([]byte, diag.Diagnostics) {
	streamsClient, diags := newStreamsClient(client)
	if diags.HasError() {
		return nil, diags
	}

	resp, err := streamsClient.GetStreamsNameWithResponse(ctx, name)
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	status := resp.StatusCode()
	switch {
	case status == http.StatusNotFound:
		return nil, nil
	case status >= 200 && status < 300:
		return resp.Body, nil
	default:
		return nil, reportUnknownError(status, resp.Body)
	}
}

// PutStream upserts a stream (PUT /api/streams/{name}).
// The caller is responsible for building a valid kbstreams.PutStreamsNameJSONRequestBody.
func PutStream(ctx context.Context, client *Client, name string, body kbstreams.PutStreamsNameJSONRequestBody) diag.Diagnostics {
	streamsClient, diags := newStreamsClient(client)
	if diags.HasError() {
		return diags
	}

	resp, err := streamsClient.PutStreamsNameWithResponse(ctx, name, nil, body)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	status := resp.StatusCode()
	if status >= 200 && status < 300 {
		return nil
	}

	return reportUnknownError(status, resp.Body)
}

// PutStreamRaw upserts a stream by sending a pre-built JSON payload to
// PUT /api/streams/{name}. This is used in cases where the generated
// kbstreams union types are too awkward to construct directly from
// Terraform models.
func PutStreamRaw(ctx context.Context, client *Client, name string, body []byte) diag.Diagnostics {
	var diags diag.Diagnostics

	endpoint := client.URL
	if !strings.HasSuffix(endpoint, "/") {
		endpoint += "/"
	}

	url := endpoint + "api/streams/" + name

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(body))
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.HTTP.Do(req)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	status := resp.StatusCode
	if status >= 200 && status < 300 {
		return diags
	}

	return reportUnknownError(status, respBody)
}

// DeleteStream deletes a stream (DELETE /api/streams/{name}).
// 404 is treated as success.
//
// We intentionally avoid the generated client here because the request body
// for this endpoint is defined as "undefined" in Kibana's schema. Some
// generated clients may serialize a nil body as JSON "null", which then
// fails validation with 'expected undefined, received null'. Using the raw
// HTTP client guarantees we send no body at all.
func DeleteStream(ctx context.Context, client *Client, name string) diag.Diagnostics {
	var diags diag.Diagnostics

	endpoint := client.URL
	if !strings.HasSuffix(endpoint, "/") {
		endpoint += "/"
	}

	url := endpoint + "api/streams/" + name

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	resp, err := client.HTTP.Do(req)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	switch resp.StatusCode {
	case http.StatusNotFound:
		return diags
	case http.StatusOK, http.StatusNoContent:
		return diags
	case http.StatusBadRequest:
		// Some Streams deployments validate the DELETE request body as
		// "undefined" and will respond with a 400 "Expected undefined,
		// received null" even though no body was sent. Treat this specific
		// validation error as a successful delete to keep Terraform idempotent.
		if bytes.Contains(respBody, []byte("Expected undefined, received null")) {
			return diags
		}
		return reportUnknownError(resp.StatusCode, respBody)
	default:
		return reportUnknownError(resp.StatusCode, respBody)
	}
}

// GetStreamIngestJSON reads ingest settings (GET /api/streams/{name}/_ingest).
// Returns nil, nil on 404.
func GetStreamIngestJSON(ctx context.Context, client *Client, name string) ([]byte, diag.Diagnostics) {
	streamsClient, diags := newStreamsClient(client)
	if diags.HasError() {
		return nil, diags
	}

	resp, err := streamsClient.GetStreamsNameIngestWithResponse(ctx, name)
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	status := resp.StatusCode()
	switch {
	case status == http.StatusNotFound:
		// No ingest definition for this stream.
		return nil, nil
	case status == http.StatusBadRequest:
		// For non‑ingest streams (e.g. group streams) Kibana returns 400
		// "Stream is not an ingest stream". In that case we simply omit the
		// ingest block from Terraform state instead of treating it as an error.
		return nil, nil
	case status >= 200 && status < 300:
		return resp.Body, nil
	default:
		return nil, reportUnknownError(status, resp.Body)
	}
}

// PutStreamIngest upserts ingest settings (PUT /api/streams/{name}/_ingest).
func PutStreamIngest(ctx context.Context, client *Client, name string, body kbstreams.PutStreamsNameIngestJSONRequestBody) diag.Diagnostics {
	streamsClient, diags := newStreamsClient(client)
	if diags.HasError() {
		return diags
	}

	resp, err := streamsClient.PutStreamsNameIngestWithResponse(ctx, name, nil, body)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	status := resp.StatusCode()
	if status >= 200 && status < 300 {
		return nil
	}

	return reportUnknownError(status, resp.Body)
}

// GetStreamGroupJSON reads group settings (GET /api/streams/{name}/_group).
// Returns nil, nil on 404.
func GetStreamGroupJSON(ctx context.Context, client *Client, name string) ([]byte, diag.Diagnostics) {
	streamsClient, diags := newStreamsClient(client)
	if diags.HasError() {
		return nil, diags
	}

	resp, err := streamsClient.GetStreamsNameGroupWithResponse(ctx, name)
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	status := resp.StatusCode()
	switch {
	case status == http.StatusNotFound:
		return nil, nil
	case status >= 200 && status < 300:
		return resp.Body, nil
	default:
		return nil, reportUnknownError(status, resp.Body)
	}
}

// PutStreamGroup upserts group settings (PUT /api/streams/{name}/_group).
func PutStreamGroup(ctx context.Context, client *Client, name string, body kbstreams.PutStreamsNameGroupJSONRequestBody) diag.Diagnostics {
	streamsClient, diags := newStreamsClient(client)
	if diags.HasError() {
		return diags
	}

	resp, err := streamsClient.PutStreamsNameGroupWithResponse(ctx, name, nil, body)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	status := resp.StatusCode()
	if status >= 200 && status < 300 {
		return nil
	}

	return reportUnknownError(status, resp.Body)
}

// EnableStreams enables Streams (POST /api/streams/_enable).
func EnableStreams(ctx context.Context, client *Client) diag.Diagnostics {
	streamsClient, diags := newStreamsClient(client)
	if diags.HasError() {
		return diags
	}

	resp, err := streamsClient.PostStreamsEnableWithResponse(ctx, nil)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	status := resp.StatusCode()
	if status >= 200 && status < 300 {
		return nil
	}

	return reportUnknownError(status, resp.Body)
}

// DisableStreams disables Streams (POST /api/streams/_disable).
func DisableStreams(ctx context.Context, client *Client) diag.Diagnostics {
	streamsClient, diags := newStreamsClient(client)
	if diags.HasError() {
		return diags
	}

	resp, err := streamsClient.PostStreamsDisableWithResponse(ctx, nil)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	status := resp.StatusCode()
	if status >= 200 && status < 300 {
		return nil
	}

	return reportUnknownError(status, resp.Body)
}
