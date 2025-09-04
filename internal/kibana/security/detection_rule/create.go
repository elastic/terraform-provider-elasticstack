package detection_rule

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// create handles the creation of a new detection rule.
func (r *Resource) create(ctx context.Context, plan *tfsdk.Plan, state *tfsdk.State, diags *diag.Diagnostics) {
	var planModel DetectionRuleModel
	diags.Append(plan.Get(ctx, &planModel)...)
	if diags.HasError() {
		return
	}

	kibanaClient, err := r.client.GetKibanaOapiClient()
	if err != nil {
		diags.AddError("Unable to get Kibana client", err.Error())
		return
	}

	// Build the create request
	createRequest, createDiags := r.buildCreateRequest(ctx, planModel)
	diags.Append(createDiags...)
	if diags.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating detection rule", map[string]interface{}{
		"name": planModel.Name.ValueString(),
		"type": planModel.Type.ValueString(),
	})

	// Create the rule
	spaceID := planModel.SpaceID.ValueString()
	if spaceID == "" {
		spaceID = "default"
	}

	resp, err := kibanaClient.API.CreateRuleWithResponse(ctx, createRequest,
		func(ctx context.Context, req *http.Request) error {
			// Add space ID header if not default space
			if spaceID != "default" {
				req.Header.Set("kbn-space-id", spaceID)
			}
			return nil
		},
	)
	if err != nil {
		diags.AddError("Error creating detection rule", fmt.Sprintf("Request failed: %s", err))
		return
	}

	if resp.StatusCode() != http.StatusOK {
		body := "unknown error"
		if resp.Body != nil {
			body = string(resp.Body)
		}
		diags.AddError("Error creating detection rule", fmt.Sprintf("API returned status %d: %s", resp.StatusCode(), body))
		return
	}

	// Parse the response
	if resp.JSON200 == nil {
		diags.AddError("Error creating detection rule", "Empty response from API")
		return
	}

	// Convert response to model
	stateModel, convertDiags := r.convertAPIResponseToModel(ctx, *resp.JSON200, spaceID)
	diags.Append(convertDiags...)
	if diags.HasError() {
		return
	}

	tflog.Debug(ctx, "Successfully created detection rule", map[string]interface{}{
		"id":      stateModel.ID.ValueString(),
		"rule_id": stateModel.RuleID.ValueString(),
	})

	// Set the state
	diags.Append(state.Set(ctx, stateModel)...)
}

// buildCreateRequest builds the API request for creating a detection rule.
func (r *Resource) buildCreateRequest(ctx context.Context, model DetectionRuleModel) (kbapi.CreateRuleJSONRequestBody, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Generate rule_id if not provided
	ruleID := model.RuleID.ValueString()
	if ruleID == "" {
		ruleID = uuid.New().String()
	}

	// Build the base request structure
	var request kbapi.SecurityDetectionsAPIRuleCreateProps

	// Set fields based on rule type
	ruleType := model.Type.ValueString()

	switch ruleType {
	case "query":
		queryRule := kbapi.SecurityDetectionsAPIQueryRuleCreateProps{
			Type:        kbapi.SecurityDetectionsAPIQueryRuleCreatePropsTypeQuery,
			Name:        model.Name.ValueString(),
			Description: model.Description.ValueString(),
			RiskScore:   int(model.RiskScore.ValueInt64()),
			Severity:    kbapi.SecurityDetectionsAPISeverity(model.Severity.ValueString()),
		}

		// Set optional fields
		if !model.Query.IsNull() && model.Query.ValueString() != "" {
			queryStr := model.Query.ValueString()
			queryRule.Query = &queryStr
		}

		if !model.Language.IsNull() && model.Language.ValueString() != "" {
			lang := kbapi.SecurityDetectionsAPIKqlQueryLanguage(model.Language.ValueString())
			queryRule.Language = &lang
		}

		if !model.Enabled.IsNull() {
			enabled := model.Enabled.ValueBool()
			queryRule.Enabled = &enabled
		}

		if ruleID != "" {
			queryRule.RuleId = &ruleID
		}

		// Use the FromQueryRuleCreateProps method to set the union
		err := request.FromSecurityDetectionsAPIQueryRuleCreateProps(queryRule)
		if err != nil {
			diags.AddError("Error building query rule request", err.Error())
			return request, diags
		}

	case "eql":
		eqlRule := kbapi.SecurityDetectionsAPIEqlRuleCreateProps{
			Type:        kbapi.SecurityDetectionsAPIEqlRuleCreatePropsTypeEql,
			Language:    kbapi.SecurityDetectionsAPIEqlQueryLanguageEql,
			Name:        model.Name.ValueString(),
			Description: model.Description.ValueString(),
			RiskScore:   int(model.RiskScore.ValueInt64()),
			Severity:    kbapi.SecurityDetectionsAPISeverity(model.Severity.ValueString()),
		}

		if !model.Query.IsNull() && model.Query.ValueString() != "" {
			queryStr := model.Query.ValueString()
			eqlRule.Query = queryStr
		}

		if !model.Enabled.IsNull() {
			enabled := model.Enabled.ValueBool()
			eqlRule.Enabled = &enabled
		}

		if ruleID != "" {
			eqlRule.RuleId = &ruleID
		}

		// Use the FromEqlRuleCreateProps method to set the union
		err := request.FromSecurityDetectionsAPIEqlRuleCreateProps(eqlRule)
		if err != nil {
			diags.AddError("Error building EQL rule request", err.Error())
			return request, diags
		}

	default:
		diags.AddError("Unsupported rule type", fmt.Sprintf("Rule type '%s' is not yet supported", ruleType))
		return request, diags
	}

	return request, diags
}

// convertAPIResponseToModel converts the API response to the Terraform model.
func (r *Resource) convertAPIResponseToModel(ctx context.Context, response interface{}, spaceID string) (DetectionRuleModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	var model DetectionRuleModel

	// Convert response to JSON for easier processing
	responseBytes, err := json.Marshal(response)
	if err != nil {
		diags.AddError("Error processing API response", fmt.Sprintf("Failed to marshal response: %s", err))
		return model, diags
	}

	// Parse as generic map for flexible handling
	var responseMap map[string]interface{}
	if err := json.Unmarshal(responseBytes, &responseMap); err != nil {
		diags.AddError("Error processing API response", fmt.Sprintf("Failed to unmarshal response: %s", err))
		return model, diags
	}

	// Extract common fields
	if id, ok := responseMap["id"].(string); ok {
		model.ID = types.StringValue(id)
	}
	if ruleId, ok := responseMap["rule_id"].(string); ok {
		model.RuleID = types.StringValue(ruleId)
	}
	if name, ok := responseMap["name"].(string); ok {
		model.Name = types.StringValue(name)
	}
	if description, ok := responseMap["description"].(string); ok {
		model.Description = types.StringValue(description)
	}
	if ruleType, ok := responseMap["type"].(string); ok {
		model.Type = types.StringValue(ruleType)
	}
	if enabled, ok := responseMap["enabled"].(bool); ok {
		model.Enabled = types.BoolValue(enabled)
	}
	if riskScore, ok := responseMap["risk_score"].(float64); ok {
		model.RiskScore = types.Int64Value(int64(riskScore))
	}
	if severity, ok := responseMap["severity"].(string); ok {
		model.Severity = types.StringValue(severity)
	}

	// Set space ID
	model.SpaceID = types.StringValue(spaceID)

	// Set computed fields
	if createdAt, ok := responseMap["created_at"].(string); ok {
		model.CreatedAt = types.StringValue(createdAt)
	}
	if createdBy, ok := responseMap["created_by"].(string); ok {
		model.CreatedBy = types.StringValue(createdBy)
	}
	if updatedAt, ok := responseMap["updated_at"].(string); ok {
		model.UpdatedAt = types.StringValue(updatedAt)
	}
	if updatedBy, ok := responseMap["updated_by"].(string); ok {
		model.UpdatedBy = types.StringValue(updatedBy)
	}
	if version, ok := responseMap["version"].(float64); ok {
		model.Version = types.Int64Value(int64(version))
	}
	if revision, ok := responseMap["revision"].(float64); ok {
		model.Revision = types.Int64Value(int64(revision))
	}

	// Handle rule-specific fields
	if query, ok := responseMap["query"].(string); ok {
		model.Query = types.StringValue(query)
	}
	if language, ok := responseMap["language"].(string); ok {
		model.Language = types.StringValue(language)
	}

	// Handle arrays - for now, set empty defaults if not present
	model.Tags = types.ListNull(types.StringType)
	model.References = types.ListNull(types.StringType)
	model.FalsePositives = types.ListNull(types.StringType)
	model.Author = types.ListNull(types.StringType)
	model.Index = types.ListNull(types.StringType)

	// Set defaults for other fields
	if model.Interval.IsNull() {
		model.Interval = types.StringValue("5m")
	}
	if model.From.IsNull() {
		model.From = types.StringValue("now-6m")
	}
	if model.To.IsNull() {
		model.To = types.StringValue("now")
	}
	if model.MaxSignals.IsNull() {
		model.MaxSignals = types.Int64Value(100)
	}

	return model, diags
}
