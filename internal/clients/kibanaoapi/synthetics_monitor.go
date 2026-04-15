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

package kibanaoapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// SyntheticsMonitorType is the monitor type discriminator.
type SyntheticsMonitorType string

const (
	SyntheticsMonitorTypeHTTP    SyntheticsMonitorType = "http"
	SyntheticsMonitorTypeTCP     SyntheticsMonitorType = "tcp"
	SyntheticsMonitorTypeICMP    SyntheticsMonitorType = "icmp"
	SyntheticsMonitorTypeBrowser SyntheticsMonitorType = "browser"
)

// SyntheticsMonitorSchedule is the schedule number+unit pair returned by the Kibana API.
type SyntheticsMonitorSchedule struct {
	Number string `json:"number"`
	Unit   string `json:"unit"`
}

// SyntheticsLocationConfig is a location object returned in the API response.
type SyntheticsLocationConfig struct {
	ID               string  `json:"id"`
	Label            string  `json:"label"`
	IsServiceManaged bool    `json:"isServiceManaged"`
	Geo              *GeoPos `json:"geo,omitempty"`
}

// GeoPos holds geographic coordinates.
type GeoPos struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

// SyntheticsMonitorAlertStatus holds the per-channel alert enable flag.
type SyntheticsMonitorAlertStatus struct {
	Enabled *bool `json:"enabled,omitempty"`
}

// SyntheticsMonitorAlert holds the alert configuration for a monitor.
type SyntheticsMonitorAlert struct {
	Status *SyntheticsMonitorAlertStatus `json:"status,omitempty"`
	TLS    *SyntheticsMonitorAlertStatus `json:"tls,omitempty"`
}

// SyntheticsMonitor is the wire representation of a Kibana synthetics monitor as returned
// by GET /api/synthetics/monitors/{id} and by POST/PUT /api/synthetics/monitors[/{id}].
//
// All fields use the same JSON keys as the legacy kbapi.SyntheticsMonitor so that
// existing mappers in schema.go can target this type with only import changes.
type SyntheticsMonitor struct {
	// Common identity fields
	ID       string                `json:"id"`
	ConfigID string                `json:"config_id"`
	Name     string                `json:"name"`
	Type     SyntheticsMonitorType `json:"type"`

	// Common config fields
	Namespace       string                     `json:"namespace"`
	Enabled         *bool                      `json:"enabled,omitempty"`
	Schedule        *SyntheticsMonitorSchedule `json:"schedule,omitempty"`
	Locations       []SyntheticsLocationConfig `json:"locations,omitempty"`
	Tags            []string                   `json:"tags,omitempty"`
	Labels          map[string]string          `json:"labels,omitempty"`
	Alert           *SyntheticsMonitorAlert    `json:"alert,omitempty"`
	APMServiceName  string                     `json:"service.name,omitempty"`
	Timeout         json.Number                `json:"timeout,omitempty"`
	Params          map[string]any             `json:"params,omitempty"`
	RetestOnFailure *bool                      `json:"retest_on_failure,omitempty"`

	// HTTP-specific fields
	URL          string         `json:"url,omitempty"`
	Mode         string         `json:"mode"`
	MaxRedirects string         `json:"max_redirects"`
	Ipv4         *bool          `json:"ipv4,omitempty"`
	Ipv6         *bool          `json:"ipv6,omitempty"`
	Username     string         `json:"username,omitempty"`
	Password     string         `json:"password,omitempty"`
	ProxyHeaders map[string]any `json:"proxy_headers,omitempty"`
	// Response and Check are preserved as raw JSON to match existing behavior.
	Response map[string]any `json:"response,omitempty"`
	Check    map[string]any `json:"check,omitempty"`

	// HTTP and TCP shared fields
	ProxyURL                  string   `json:"proxy_url,omitempty"`
	SslVerificationMode       string   `json:"ssl.verification_mode"`
	SslSupportedProtocols     []string `json:"ssl.supported_protocols"`
	SslCertificateAuthorities []string `json:"ssl.certificate_authorities,omitempty"`
	SslCertificate            string   `json:"ssl.certificate,omitempty"`
	SslKey                    string   `json:"ssl.key,omitempty"`
	SslKeyPassphrase          string   `json:"ssl.key_passphrase,omitempty"`

	// TCP and ICMP shared field
	Host string `json:"host,omitempty"`

	// TCP-specific fields
	ProxyUseLocalResolver *bool  `json:"proxy_use_local_resolver,omitempty"`
	CheckSend             string `json:"check.send,omitempty"`
	CheckReceive          string `json:"check.receive,omitempty"`

	// ICMP-specific field
	Wait json.Number `json:"wait,omitempty"`

	// Browser-specific fields
	Screenshots       string         `json:"screenshots,omitempty"`
	IgnoreHTTPSErrors *bool          `json:"ignore_https_errors,omitempty"`
	InlineScript      string         `json:"inline_script"`
	SyntheticsArgs    []string       `json:"synthetics_args,omitempty"`
	PlaywrightOptions map[string]any `json:"playwright_options,omitempty"`
}

// SyntheticsMonitorRequest is the wire body sent for POST and PUT monitor requests.
// It embeds all the fields flat (as the legacy client did via APIRequest).
type SyntheticsMonitorRequest struct {
	// type discriminator
	Type SyntheticsMonitorType `json:"type"`

	// Common config
	Name             string                  `json:"name"`
	Schedule         int64                   `json:"schedule,omitempty"`
	Locations        []string                `json:"locations,omitempty"`
	PrivateLocations []string                `json:"private_locations,omitempty"`
	Enabled          *bool                   `json:"enabled,omitempty"`
	Tags             []string                `json:"tags,omitempty"`
	Labels           map[string]string       `json:"labels"`
	Alert            *SyntheticsMonitorAlert `json:"alert,omitempty"`
	APMServiceName   string                  `json:"service.name,omitempty"`
	TimeoutSeconds   int                     `json:"timeout,omitempty"`
	Namespace        string                  `json:"namespace,omitempty"`
	Params           map[string]any          `json:"params,omitempty"`
	RetestOnFailure  *bool                   `json:"retest_on_failure,omitempty"`

	// HTTP-specific
	URL          string               `json:"url,omitempty"`
	Ssl          *SyntheticsSSLConfig `json:"ssl,omitempty"`
	MaxRedirects string               `json:"max_redirects,omitempty"`
	Mode         string               `json:"mode,omitempty"`
	Ipv4         *bool                `json:"ipv4,omitempty"`
	Ipv6         *bool                `json:"ipv6,omitempty"`
	Username     string               `json:"username,omitempty"`
	Password     string               `json:"password,omitempty"`
	ProxyHeader  map[string]any       `json:"proxy_headers,omitempty"`
	ProxyURL     string               `json:"proxy_url,omitempty"`
	Response     map[string]any       `json:"response,omitempty"`
	Check        map[string]any       `json:"check,omitempty"`

	// TCP-specific
	Host                  string `json:"host,omitempty"`
	CheckSend             string `json:"check.send,omitempty"`
	CheckReceive          string `json:"check.receive,omitempty"`
	ProxyUseLocalResolver *bool  `json:"proxy_use_local_resolver,omitempty"`

	// ICMP-specific
	Wait string `json:"wait,omitempty"`

	// Browser-specific
	InlineScript      string         `json:"inline_script,omitempty"`
	Screenshots       string         `json:"screenshots,omitempty"`
	SyntheticsArgs    []string       `json:"synthetics_args,omitempty"`
	IgnoreHTTPSErrors *bool          `json:"ignore_https_errors,omitempty"`
	PlaywrightOptions map[string]any `json:"playwright_options,omitempty"`
}

// SyntheticsSSLConfig holds SSL/TLS configuration for monitors.
type SyntheticsSSLConfig struct {
	VerificationMode       string   `json:"verification_mode,omitempty"`
	SupportedProtocols     []string `json:"supported_protocols,omitempty"`
	CertificateAuthorities []string `json:"certificate_authorities,omitempty"`
	Certificate            string   `json:"certificate,omitempty"`
	Key                    string   `json:"key,omitempty"`
	KeyPassphrase          string   `json:"key_passphrase,omitempty"`
}

// CreateMonitor creates a new synthetics monitor via POST /api/synthetics/monitors.
// The raw response body is unmarshaled into a SyntheticsMonitor — Kibana returns
// the full monitor object in the POST response body, not just warnings.
func CreateMonitor(ctx context.Context, client *Client, spaceID string, req SyntheticsMonitorRequest) (*SyntheticsMonitor, diag.Diagnostics) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	resp, err := client.API.PostSyntheticMonitorsWithBodyWithResponse(
		ctx, "application/json", bytes.NewReader(body),
		SpaceAwarePathRequestEditor(spaceID),
	)
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		var monitor SyntheticsMonitor
		if jsonErr := json.Unmarshal(resp.Body, &monitor); jsonErr != nil {
			return nil, diagutil.FrameworkDiagFromError(jsonErr)
		}
		return &monitor, nil
	default:
		return nil, reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// GetMonitor reads a synthetics monitor by ID via GET /api/synthetics/monitors/{id}.
// Returns nil, nil when the monitor is not found (404).
func GetMonitor(ctx context.Context, client *Client, spaceID string, monitorID string) (*SyntheticsMonitor, diag.Diagnostics) {
	resp, err := client.API.GetSyntheticMonitorWithResponse(
		ctx, monitorID,
		SpaceAwarePathRequestEditor(spaceID),
	)
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		var monitor SyntheticsMonitor
		if jsonErr := json.Unmarshal(resp.Body, &monitor); jsonErr != nil {
			return nil, diagutil.FrameworkDiagFromError(jsonErr)
		}
		return &monitor, nil
	case http.StatusNotFound:
		return nil, nil
	default:
		return nil, reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// UpdateMonitor updates a synthetics monitor via PUT /api/synthetics/monitors/{id}.
// The raw response body is unmarshaled into a SyntheticsMonitor.
func UpdateMonitor(ctx context.Context, client *Client, spaceID string, monitorID string, req SyntheticsMonitorRequest) (*SyntheticsMonitor, diag.Diagnostics) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	resp, err := client.API.PutSyntheticMonitorWithBodyWithResponse(
		ctx, monitorID, "application/json", bytes.NewReader(body),
		SpaceAwarePathRequestEditor(spaceID),
	)
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		var monitor SyntheticsMonitor
		if jsonErr := json.Unmarshal(resp.Body, &monitor); jsonErr != nil {
			return nil, diagutil.FrameworkDiagFromError(jsonErr)
		}
		return &monitor, nil
	default:
		return nil, reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// monitorDeleteStatus represents a single item in the Kibana bulk-delete response body.
type monitorDeleteStatus struct {
	ID      string `json:"id"`
	Deleted bool   `json:"deleted"`
}

// DeleteMonitor deletes a synthetics monitor via DELETE /api/synthetics/monitors.
// It uses the bulk-delete endpoint (body: {"ids": [monitorID]}) which is supported
// across all Kibana versions that have the synthetics monitor resource (8.14+).
// The per-id endpoint (DELETE /api/synthetics/monitors/{id}) was not available in
// older Kibana versions and would return 404, causing silent deletion failures.
// The response body is parsed to detect per-item failures reported as HTTP 200 with
// deleted=false (e.g., when the monitor is in use or an internal error occurs).
func DeleteMonitor(ctx context.Context, client *Client, spaceID string, monitorID string) diag.Diagnostics {
	body, err := json.Marshal(map[string]any{"ids": []string{monitorID}})
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	path := BuildSpaceAwarePath(spaceID, "/api/synthetics/monitors")
	url := strings.TrimRight(client.URL, "/") + path

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, bytes.NewReader(body))
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.HTTP.Do(req)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, _ := io.ReadAll(resp.Body)

	switch resp.StatusCode {
	case http.StatusNoContent, http.StatusNotFound:
		return nil
	case http.StatusOK:
		// Parse the per-item result array to detect partial failures.
		var results []monitorDeleteStatus
		if jsonErr := json.Unmarshal(respBody, &results); jsonErr != nil {
			// If the response body cannot be parsed, assume success — the HTTP 200
			// indicates the request was accepted by the server.
			return nil
		}
		for _, r := range results {
			if r.ID == monitorID && !r.Deleted {
				return diagutil.FrameworkDiagFromError(fmt.Errorf("monitor %s was not deleted by Kibana", monitorID))
			}
		}
		return nil
	default:
		return diagutil.FrameworkDiagFromError(fmt.Errorf("unexpected status %d deleting monitor %s: %s", resp.StatusCode, monitorID, string(respBody)))
	}
}
