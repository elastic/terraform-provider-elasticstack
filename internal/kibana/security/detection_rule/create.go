package detection_rule

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
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

	// Debug the request payload
	requestJSON, _ := json.Marshal(createRequest)
	tflog.Debug(ctx, "Creating detection rule", map[string]interface{}{
		"name":    planModel.Name.ValueString(),
		"type":    planModel.Type.ValueString(),
		"request": string(requestJSON),
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
	stateModel, convertDiags := r.convertAPIResponseToModel(ctx, *resp.JSON200, spaceID, &planModel)
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

	// Set fields based on rule type
	ruleType := model.Type.ValueString()

	// Build the request directly as JSON to avoid union type issues
	var requestData interface{}

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

		// Set list fields from plan
		if !model.Tags.IsNull() && len(model.Tags.Elements()) > 0 {
			tags := make([]string, 0, len(model.Tags.Elements()))
			for _, elem := range model.Tags.Elements() {
				if str, ok := elem.(types.String); ok && !str.IsNull() {
					tags = append(tags, str.ValueString())
				}
			}
			if len(tags) > 0 {
				queryRule.Tags = &tags
			}
		}

		if !model.References.IsNull() && len(model.References.Elements()) > 0 {
			references := make([]string, 0, len(model.References.Elements()))
			for _, elem := range model.References.Elements() {
				if str, ok := elem.(types.String); ok && !str.IsNull() {
					references = append(references, str.ValueString())
				}
			}
			if len(references) > 0 {
				queryRule.References = &references
			}
		}

		if !model.FalsePositives.IsNull() && len(model.FalsePositives.Elements()) > 0 {
			falsePositives := make([]string, 0, len(model.FalsePositives.Elements()))
			for _, elem := range model.FalsePositives.Elements() {
				if str, ok := elem.(types.String); ok && !str.IsNull() {
					falsePositives = append(falsePositives, str.ValueString())
				}
			}
			if len(falsePositives) > 0 {
				queryRule.FalsePositives = &falsePositives
			}
		}

		if !model.Author.IsNull() && len(model.Author.Elements()) > 0 {
			author := make([]string, 0, len(model.Author.Elements()))
			for _, elem := range model.Author.Elements() {
				if str, ok := elem.(types.String); ok && !str.IsNull() {
					author = append(author, str.ValueString())
				}
			}
			if len(author) > 0 {
				queryRule.Author = &author
			}
		}

		requestData = queryRule

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

		// Set list fields from plan
		if !model.Tags.IsNull() && len(model.Tags.Elements()) > 0 {
			tags := make([]string, 0, len(model.Tags.Elements()))
			for _, elem := range model.Tags.Elements() {
				if str, ok := elem.(types.String); ok && !str.IsNull() {
					tags = append(tags, str.ValueString())
				}
			}
			if len(tags) > 0 {
				eqlRule.Tags = &tags
			}
		}

		if !model.References.IsNull() && len(model.References.Elements()) > 0 {
			references := make([]string, 0, len(model.References.Elements()))
			for _, elem := range model.References.Elements() {
				if str, ok := elem.(types.String); ok && !str.IsNull() {
					references = append(references, str.ValueString())
				}
			}
			if len(references) > 0 {
				eqlRule.References = &references
			}
		}

		if !model.FalsePositives.IsNull() && len(model.FalsePositives.Elements()) > 0 {
			falsePositives := make([]string, 0, len(model.FalsePositives.Elements()))
			for _, elem := range model.FalsePositives.Elements() {
				if str, ok := elem.(types.String); ok && !str.IsNull() {
					falsePositives = append(falsePositives, str.ValueString())
				}
			}
			if len(falsePositives) > 0 {
				eqlRule.FalsePositives = &falsePositives
			}
		}

		if !model.Author.IsNull() && len(model.Author.Elements()) > 0 {
			author := make([]string, 0, len(model.Author.Elements()))
			for _, elem := range model.Author.Elements() {
				if str, ok := elem.(types.String); ok && !str.IsNull() {
					author = append(author, str.ValueString())
				}
			}
			if len(author) > 0 {
				eqlRule.Author = &author
			}
		}

		requestData = eqlRule

	default:
		diags.AddError("Unsupported rule type", fmt.Sprintf("Rule type '%s' is not yet supported", ruleType))
		return kbapi.SecurityDetectionsAPIRuleCreateProps{}, diags
	}

	// Marshal to JSON and then unmarshal to the union type to preserve the correct structure
	jsonData, err := json.Marshal(requestData)
	if err != nil {
		diags.AddError("Error marshaling request data", err.Error())
		return kbapi.SecurityDetectionsAPIRuleCreateProps{}, diags
	}

	var request kbapi.SecurityDetectionsAPIRuleCreateProps
	if err := json.Unmarshal(jsonData, &request); err != nil {
		diags.AddError("Error creating union type", err.Error())
		return kbapi.SecurityDetectionsAPIRuleCreateProps{}, diags
	}

	return request, diags
}

// convertAPIResponseToModel converts the API response to the Terraform model.
func (r *Resource) convertAPIResponseToModel(ctx context.Context, response interface{}, spaceID string, planModel *DetectionRuleModel) (DetectionRuleModel, diag.Diagnostics) {
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
	} else if planModel != nil && !planModel.Query.IsNull() {
		model.Query = planModel.Query
	}

	if language, ok := responseMap["language"].(string); ok {
		model.Language = types.StringValue(language)
	} else if planModel != nil && !planModel.Language.IsNull() {
		model.Language = planModel.Language
	}

	// Handle list attributes - preserve plan values or extract from response
	if tagsInterface, ok := responseMap["tags"].([]interface{}); ok {
		tags := make([]attr.Value, len(tagsInterface))
		for i, v := range tagsInterface {
			if str, ok := v.(string); ok {
				tags[i] = types.StringValue(str)
			}
		}
		model.Tags = types.ListValueMust(types.StringType, tags)
	} else if planModel != nil && !planModel.Tags.IsNull() {
		model.Tags = planModel.Tags
	} else {
		model.Tags = types.ListValueMust(types.StringType, []attr.Value{})
	}

	if referencesInterface, ok := responseMap["references"].([]interface{}); ok {
		references := make([]attr.Value, len(referencesInterface))
		for i, v := range referencesInterface {
			if str, ok := v.(string); ok {
				references[i] = types.StringValue(str)
			}
		}
		model.References = types.ListValueMust(types.StringType, references)
	} else if planModel != nil && !planModel.References.IsNull() {
		model.References = planModel.References
	} else {
		model.References = types.ListValueMust(types.StringType, []attr.Value{})
	}

	if falsePositivesInterface, ok := responseMap["false_positives"].([]interface{}); ok {
		falsePositives := make([]attr.Value, len(falsePositivesInterface))
		for i, v := range falsePositivesInterface {
			if str, ok := v.(string); ok {
				falsePositives[i] = types.StringValue(str)
			}
		}
		model.FalsePositives = types.ListValueMust(types.StringType, falsePositives)
	} else if planModel != nil && !planModel.FalsePositives.IsNull() {
		model.FalsePositives = planModel.FalsePositives
	} else {
		model.FalsePositives = types.ListValueMust(types.StringType, []attr.Value{})
	}

	if authorInterface, ok := responseMap["author"].([]interface{}); ok {
		author := make([]attr.Value, len(authorInterface))
		for i, v := range authorInterface {
			if str, ok := v.(string); ok {
				author[i] = types.StringValue(str)
			}
		}
		model.Author = types.ListValueMust(types.StringType, author)
	} else if planModel != nil && !planModel.Author.IsNull() {
		model.Author = planModel.Author
	} else {
		model.Author = types.ListValueMust(types.StringType, []attr.Value{})
	}

	if indexInterface, ok := responseMap["index"].([]interface{}); ok {
		index := make([]attr.Value, len(indexInterface))
		for i, v := range indexInterface {
			if str, ok := v.(string); ok {
				index[i] = types.StringValue(str)
			}
		}
		model.Index = types.ListValueMust(types.StringType, index)
	} else if planModel != nil && !planModel.Index.IsNull() {
		model.Index = planModel.Index
	} else {
		model.Index = types.ListNull(types.StringType)
	}

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
