package kibana_oapi

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func GetStreamJSON(ctx context.Context, client *Client, name string) ([]byte, diag.Diagnostics) {
	resp, err := client.API.GetStreamsNameWithResponse(ctx, name)
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
		if bytes.Contains(respBody, []byte("Expected undefined, received null")) {
			return diags
		}
		return reportUnknownError(resp.StatusCode, respBody)
	default:
		return reportUnknownError(resp.StatusCode, respBody)
	}
}

func GetStreamIngestJSON(ctx context.Context, client *Client, name string) ([]byte, diag.Diagnostics) {
	resp, err := client.API.GetStreamsNameIngestWithResponse(ctx, name)
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	status := resp.StatusCode()
	switch {
	case status == http.StatusNotFound:
		return nil, nil
	case status == http.StatusBadRequest:

		return nil, nil
	case status >= 200 && status < 300:
		return resp.Body, nil
	default:
		return nil, reportUnknownError(status, resp.Body)
	}
}

func PutStreamIngest(ctx context.Context, client *Client, name string, body kbapi.PutStreamsNameIngestJSONRequestBody) diag.Diagnostics {
	resp, err := client.API.PutStreamsNameIngestWithResponse(ctx, name, body)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	status := resp.StatusCode()
	if status >= 200 && status < 300 {
		return nil
	}

	return reportUnknownError(status, resp.Body)
}

func GetStreamGroupJSON(ctx context.Context, client *Client, name string) ([]byte, diag.Diagnostics) {
	resp, err := client.API.GetStreamsNameGroupWithResponse(ctx, name)
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

func PutStreamGroup(ctx context.Context, client *Client, name string, body kbapi.PutStreamsNameGroupJSONRequestBody) diag.Diagnostics {
	resp, err := client.API.PutStreamsNameGroupWithResponse(ctx, name, body)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	status := resp.StatusCode()
	if status >= 200 && status < 300 {
		return nil
	}

	return reportUnknownError(status, resp.Body)
}

func EnableStreams(ctx context.Context, client *Client) diag.Diagnostics {
	resp, err := client.API.PostStreamsEnableWithResponse(ctx)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	status := resp.StatusCode()
	if status >= 200 && status < 300 {
		return nil
	}

	return reportUnknownError(status, resp.Body)
}

func DisableStreams(ctx context.Context, client *Client) diag.Diagnostics {
	resp, err := client.API.PostStreamsDisableWithResponse(ctx)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	status := resp.StatusCode()
	if status >= 200 && status < 300 {
		return nil
	}

	return reportUnknownError(status, resp.Body)
}
