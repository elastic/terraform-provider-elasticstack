package detection_rule

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// SecurityDetectionRuleRequest represents a security detection rule creation/update request
type SecurityDetectionRuleRequest struct {
	Name              string          `json:"name"`
	Description       string          `json:"description"`
	Type              string          `json:"type"`
	Query             *string         `json:"query,omitempty"`
	Language          *string         `json:"language,omitempty"`
	Index             []string        `json:"index,omitempty"`
	Severity          string          `json:"severity"`
	Risk              int             `json:"risk_score"`
	Enabled           bool            `json:"enabled"`
	Tags              []string        `json:"tags,omitempty"`
	From              string          `json:"from"`
	To                string          `json:"to"`
	Interval          string          `json:"interval"`
	Meta              *map[string]any `json:"meta,omitempty"`
	Author            []string        `json:"author,omitempty"`
	License           *string         `json:"license,omitempty"`
	RuleNameOverride  *string         `json:"rule_name_override,omitempty"`
	TimestampOverride *string         `json:"timestamp_override,omitempty"`
	Note              *string         `json:"note,omitempty"`
	References        []string        `json:"references,omitempty"`
	FalsePositives    []string        `json:"false_positives,omitempty"`
	ExceptionsList    []any           `json:"exceptions_list,omitempty"`
	Version           int             `json:"version"`
	MaxSignals        int             `json:"max_signals"`
}

// SecurityDetectionRuleResponse represents the API response for a security detection rule
type SecurityDetectionRuleResponse struct {
	ID                string          `json:"id"`
	Name              string          `json:"name"`
	Description       string          `json:"description"`
	Type              string          `json:"type"`
	Query             *string         `json:"query,omitempty"`
	Language          *string         `json:"language,omitempty"`
	Index             []string        `json:"index,omitempty"`
	Severity          string          `json:"severity"`
	Risk              int             `json:"risk_score"`
	Enabled           bool            `json:"enabled"`
	Tags              []string        `json:"tags,omitempty"`
	From              string          `json:"from"`
	To                string          `json:"to"`
	Interval          string          `json:"interval"`
	Meta              *map[string]any `json:"meta,omitempty"`
	Author            []string        `json:"author,omitempty"`
	License           *string         `json:"license,omitempty"`
	RuleNameOverride  *string         `json:"rule_name_override,omitempty"`
	TimestampOverride *string         `json:"timestamp_override,omitempty"`
	Note              *string         `json:"note,omitempty"`
	References        []string        `json:"references,omitempty"`
	FalsePositives    []string        `json:"false_positives,omitempty"`
	ExceptionsList    []any           `json:"exceptions_list,omitempty"`
	Version           int             `json:"version"`
	MaxSignals        int             `json:"max_signals"`
	CreatedAt         string          `json:"created_at"`
	CreatedBy         string          `json:"created_by"`
	UpdatedAt         string          `json:"updated_at"`
	UpdatedBy         string          `json:"updated_by"`
}

// CreateSecurityDetectionRule creates a new security detection rule
func CreateSecurityDetectionRule(ctx context.Context, client *clients.ApiClient, spaceId string, ruleId *string, rule *SecurityDetectionRuleRequest) (*SecurityDetectionRuleResponse, diag.Diagnostics) {
	var diags diag.Diagnostics

	kbClient, err := client.GetKibanaClient()
	if err != nil {
		diags.AddError("Failed to get Kibana client", err.Error())
		return nil, diags
	}

	// Create the URL path
	path := fmt.Sprintf("/s/%s/api/detection_engine/rules", url.PathEscape(spaceId))

	// Execute the request using resty
	resp, err := kbClient.Client.R().SetBody(rule).Post(path)
	if err != nil {
		diags.AddError("Failed to execute request", err.Error())
		return nil, diags
	}

	// Handle non-2xx status codes
	if resp.StatusCode() >= 300 {
		diags.AddError(
			"API request failed",
			fmt.Sprintf("Status: %d, URL: %s, Body: %s", resp.StatusCode(), resp.Request.URL, string(resp.Body())),
		)
		return nil, diags
	}

	// Parse the response
	var result SecurityDetectionRuleResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		diags.AddError("Failed to decode response", err.Error())
		return nil, diags
	}

	return &result, diags
}

// GetSecurityDetectionRule retrieves a security detection rule by ID
func GetSecurityDetectionRule(ctx context.Context, client *clients.ApiClient, spaceId, ruleId string) (*SecurityDetectionRuleResponse, diag.Diagnostics) {
	var diags diag.Diagnostics

	kbClient, err := client.GetKibanaClient()
	if err != nil {
		diags.AddError("Failed to get Kibana client", err.Error())
		return nil, diags
	}

	// Create the URL path
	path := fmt.Sprintf("/s/%s/api/detection_engine/rules?id=%s", url.PathEscape(spaceId), url.QueryEscape(ruleId))

	// Execute the request using resty
	resp, err := kbClient.Client.R().Get(path)
	if err != nil {
		diags.AddError("Failed to execute request", err.Error())
		return nil, diags
	}

	// Handle not found
	if resp.StatusCode() == 404 {
		return nil, diags // Rule not found
	}

	// Handle other non-2xx status codes
	if resp.StatusCode() >= 300 {
		diags.AddError(
			"API request failed",
			fmt.Sprintf("Status: %d, URL: %s, Body: %s", resp.StatusCode(), resp.Request.URL, string(resp.Body())),
		)
		return nil, diags
	}

	// Parse the response
	var result SecurityDetectionRuleResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		diags.AddError("Failed to decode response", err.Error())
		return nil, diags
	}

	return &result, diags
}

// UpdateSecurityDetectionRule updates an existing security detection rule
func UpdateSecurityDetectionRule(ctx context.Context, client *clients.ApiClient, spaceId, ruleId string, rule *SecurityDetectionRuleRequest) (*SecurityDetectionRuleResponse, diag.Diagnostics) {
	var diags diag.Diagnostics

	kbClient, err := client.GetKibanaClient()
	if err != nil {
		diags.AddError("Failed to get Kibana client", err.Error())
		return nil, diags
	}

	// Create the URL path
	path := fmt.Sprintf("/s/%s/api/detection_engine/rules", url.PathEscape(spaceId))

	// Execute the request using resty
	resp, err := kbClient.Client.R().SetBody(rule).Put(path)
	if err != nil {
		diags.AddError("Failed to execute request", err.Error())
		return nil, diags
	}

	// Handle non-2xx status codes
	if resp.StatusCode() >= 300 {
		diags.AddError(
			"API request failed",
			fmt.Sprintf("Status: %d, URL: %s, Body: %s", resp.StatusCode(), resp.Request.URL, string(resp.Body())),
		)
		return nil, diags
	}

	// Parse the response
	var result SecurityDetectionRuleResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		diags.AddError("Failed to decode response", err.Error())
		return nil, diags
	}

	return &result, diags
}

// DeleteSecurityDetectionRule deletes a security detection rule by ID
func DeleteSecurityDetectionRule(ctx context.Context, client *clients.ApiClient, spaceId, ruleId string) diag.Diagnostics {
	var diags diag.Diagnostics

	kbClient, err := client.GetKibanaClient()
	if err != nil {
		diags.AddError("Failed to get Kibana client", err.Error())
		return diags
	}

	// Create the URL path
	path := fmt.Sprintf("/s/%s/api/detection_engine/rules?id=%s", url.PathEscape(spaceId), url.QueryEscape(ruleId))

	// Execute the request using resty
	resp, err := kbClient.Client.R().Delete(path)
	if err != nil {
		diags.AddError("Failed to execute request", err.Error())
		return diags
	}

	// Handle not found (rule might already be deleted)
	if resp.StatusCode() == 404 {
		return diags // Already deleted, no error
	}

	// Handle other non-2xx status codes
	if resp.StatusCode() >= 300 {
		diags.AddError(
			"API request failed",
			fmt.Sprintf("Status: %d, URL: %s, Body: %s", resp.StatusCode(), resp.Request.URL, string(resp.Body())),
		)
		return diags
	}

	return diags
}
