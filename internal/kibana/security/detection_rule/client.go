package detection_rule

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// CreateSecurityDetectionRule creates a new security detection rule using the generated API client
func CreateSecurityDetectionRule(ctx context.Context, client *clients.ApiClient, spaceId string, rule *SecurityDetectionRuleRequest) (*SecurityDetectionRuleResponse, diag.Diagnostics) {
	var diags diag.Diagnostics

	kbClient, err := client.GetKibanaClient()
	if err != nil {
		diags.AddError("Failed to get Kibana client", err.Error())
		return nil, diags
	}

	// Create the generated API client
	genClient, err := kbapi.NewClientWithResponses(kbClient.Client.BaseURL, kbapi.WithHTTPClient(kbClient.Client.GetClient()))
	if err != nil {
		diags.AddError("Failed to create generated client", err.Error())
		return nil, diags
	}

	// Convert our request to the generated API types
	createProps := kbapi.SecurityDetectionsAPIRuleCreateProps{}
	
	// Create a QueryRuleCreateProps for simplicity (we can extend this later for other rule types)
	enabled := kbapi.SecurityDetectionsAPIIsRuleEnabled(rule.Enabled)
	from := kbapi.SecurityDetectionsAPIRuleIntervalFrom(rule.From)
	to := kbapi.SecurityDetectionsAPIRuleIntervalTo(rule.To)
	interval := kbapi.SecurityDetectionsAPIRuleInterval(rule.Interval)

	queryRuleProps := kbapi.SecurityDetectionsAPIQueryRuleCreateProps{
		Name:        rule.Name,
		Description: kbapi.SecurityDetectionsAPIRuleDescription(rule.Description),
		Type:        kbapi.SecurityDetectionsAPIQueryRuleCreatePropsType(rule.Type),
		Severity:    kbapi.SecurityDetectionsAPISeverity(rule.Severity),
		RiskScore:   rule.Risk,
		Enabled:     &enabled,
		From:        &from,
		To:          &to,
		Interval:    &interval,
		MaxSignals:  &rule.MaxSignals,
		Version:     &rule.Version,
	}

	// Set optional fields
	if rule.Query != nil {
		query := kbapi.SecurityDetectionsAPIRuleQuery(*rule.Query)
		queryRuleProps.Query = &query
	}
	if rule.Language != nil {
		language := kbapi.SecurityDetectionsAPIKqlQueryLanguage(*rule.Language)
		queryRuleProps.Language = &language
	}
	if len(rule.Index) > 0 {
		indexArray := make(kbapi.SecurityDetectionsAPIIndexPatternArray, len(rule.Index))
		for i, idx := range rule.Index {
			indexArray[i] = idx
		}
		queryRuleProps.Index = &indexArray
	}
	if len(rule.Tags) > 0 {
		tagArray := make(kbapi.SecurityDetectionsAPIRuleTagArray, len(rule.Tags))
		for i, tag := range rule.Tags {
			tagArray[i] = tag
		}
		queryRuleProps.Tags = &tagArray
	}
	if len(rule.Author) > 0 {
		authorArray := make(kbapi.SecurityDetectionsAPIRuleAuthorArray, len(rule.Author))
		for i, author := range rule.Author {
			authorArray[i] = author
		}
		queryRuleProps.Author = &authorArray
	}
	if rule.License != nil {
		license := kbapi.SecurityDetectionsAPIRuleLicense(*rule.License)
		queryRuleProps.License = &license
	}
	if rule.RuleNameOverride != nil {
		override := kbapi.SecurityDetectionsAPIRuleNameOverride(*rule.RuleNameOverride)
		queryRuleProps.RuleNameOverride = &override
	}
	if rule.TimestampOverride != nil {
		timestampOverride := kbapi.SecurityDetectionsAPITimestampOverride(*rule.TimestampOverride)
		queryRuleProps.TimestampOverride = &timestampOverride
	}
	if rule.Note != nil {
		note := kbapi.SecurityDetectionsAPIInvestigationGuide(*rule.Note)
		queryRuleProps.Note = &note
	}
	if len(rule.References) > 0 {
		refArray := make(kbapi.SecurityDetectionsAPIRuleReferenceArray, len(rule.References))
		for i, ref := range rule.References {
			refArray[i] = ref
		}
		queryRuleProps.References = &refArray
	}
	if len(rule.FalsePositives) > 0 {
		fpArray := make(kbapi.SecurityDetectionsAPIRuleFalsePositiveArray, len(rule.FalsePositives))
		for i, fp := range rule.FalsePositives {
			fpArray[i] = fp
		}
		queryRuleProps.FalsePositives = &fpArray
	}

	// Set the query rule props in the union
	if err := createProps.FromSecurityDetectionsAPIQueryRuleCreateProps(queryRuleProps); err != nil {
		diags.AddError("Failed to create request body", err.Error())
		return nil, diags
	}

	// Call the generated API
	resp, err := genClient.CreateRuleWithResponse(ctx, kbapi.SpaceId(spaceId), createProps)
	if err != nil {
		diags.AddError("Failed to execute request", err.Error())
		return nil, diags
	}

	// Check for API errors
	if resp.StatusCode() >= 300 {
		diags.AddError(
			"API request failed",
			fmt.Sprintf("Status: %d, Body: %s", resp.StatusCode(), string(resp.Body)),
		)
		return nil, diags
	}

	// Parse the response - it's a union type so we need to convert it
	if resp.JSON200 == nil {
		diags.AddError("Unexpected response", "Expected JSON response but got nil")
		return nil, diags
	}

	ruleResponse, err := resp.JSON200.AsSecurityDetectionsAPIQueryRule()
	if err != nil {
		diags.AddError("Failed to parse response", err.Error())
		return nil, diags
	}

	// Convert the response to our internal type
	result := &SecurityDetectionRuleResponse{
		ID:          ruleResponse.Id.String(),
		Name:        ruleResponse.Name,
		Description: string(ruleResponse.Description),
		Type:        string(ruleResponse.Type),
		Severity:    string(ruleResponse.Severity),
		Risk:        ruleResponse.RiskScore,
		Enabled:     bool(ruleResponse.Enabled),
		From:        string(ruleResponse.From),
		To:          string(ruleResponse.To),
		Interval:    string(ruleResponse.Interval),
		Version:     ruleResponse.Version,
		MaxSignals:  ruleResponse.MaxSignals,
		CreatedAt:   ruleResponse.CreatedAt.Format("2006-01-02T15:04:05.000Z"),
		CreatedBy:   ruleResponse.CreatedBy,
		UpdatedAt:   ruleResponse.UpdatedAt.Format("2006-01-02T15:04:05.000Z"),
		UpdatedBy:   ruleResponse.UpdatedBy,
	}

	// Set optional fields
	queryStr := string(ruleResponse.Query)
	result.Query = &queryStr

	langStr := string(ruleResponse.Language)
	result.Language = &langStr

	if ruleResponse.Index != nil {
		result.Index = make([]string, len(*ruleResponse.Index))
		for i, idx := range *ruleResponse.Index {
			result.Index[i] = idx
		}
	}
	result.Tags = make([]string, len(ruleResponse.Tags))
	for i, tag := range ruleResponse.Tags {
		result.Tags[i] = tag
	}
	result.Author = make([]string, len(ruleResponse.Author))
	for i, author := range ruleResponse.Author {
		result.Author[i] = author
	}
	if ruleResponse.License != nil {
		licenseStr := string(*ruleResponse.License)
		result.License = &licenseStr
	}
	if ruleResponse.RuleNameOverride != nil {
		overrideStr := string(*ruleResponse.RuleNameOverride)
		result.RuleNameOverride = &overrideStr
	}
	if ruleResponse.TimestampOverride != nil {
		timestampStr := string(*ruleResponse.TimestampOverride)
		result.TimestampOverride = &timestampStr
	}
	if ruleResponse.Note != nil {
		noteStr := string(*ruleResponse.Note)
		result.Note = &noteStr
	}
	result.References = make([]string, len(ruleResponse.References))
	for i, ref := range ruleResponse.References {
		result.References[i] = ref
	}
	result.FalsePositives = make([]string, len(ruleResponse.FalsePositives))
	for i, fp := range ruleResponse.FalsePositives {
		result.FalsePositives[i] = fp
	}

	return result, diags
}

// GetSecurityDetectionRule retrieves a security detection rule by ID using the generated API client
func GetSecurityDetectionRule(ctx context.Context, client *clients.ApiClient, spaceId, ruleId string) (*SecurityDetectionRuleResponse, diag.Diagnostics) {
	var diags diag.Diagnostics

	kbClient, err := client.GetKibanaClient()
	if err != nil {
		diags.AddError("Failed to get Kibana client", err.Error())
		return nil, diags
	}

	// Create the generated API client
	genClient, err := kbapi.NewClientWithResponses(kbClient.Client.BaseURL, kbapi.WithHTTPClient(kbClient.Client.GetClient()))
	if err != nil {
		diags.AddError("Failed to create generated client", err.Error())
		return nil, diags
	}

	// Set up parameters - use rule ID for reading
	parsedId, err := uuid.Parse(ruleId)
	if err != nil {
		diags.AddError("Invalid rule ID", fmt.Sprintf("Failed to parse rule ID as UUID: %s", err.Error()))
		return nil, diags
	}
	id := kbapi.SecurityDetectionsAPIRuleObjectId(parsedId)
	params := &kbapi.ReadRuleParams{
		Id: &id,
	}

	// Call the generated API
	resp, err := genClient.ReadRuleWithResponse(ctx, kbapi.SpaceId(spaceId), params)
	if err != nil {
		diags.AddError("Failed to execute request", err.Error())
		return nil, diags
	}

	// Handle not found
	if resp.StatusCode() == 404 {
		return nil, diags // Rule not found
	}

	// Check for other API errors
	if resp.StatusCode() >= 300 {
		diags.AddError(
			"API request failed",
			fmt.Sprintf("Status: %d, Body: %s", resp.StatusCode(), string(resp.Body)),
		)
		return nil, diags
	}

	// Parse the response
	if resp.JSON200 == nil {
		diags.AddError("Unexpected response", "Expected JSON response but got nil")
		return nil, diags
	}

	ruleResponse, err := resp.JSON200.AsSecurityDetectionsAPIQueryRule()
	if err != nil {
		diags.AddError("Failed to parse response", err.Error())
		return nil, diags
	}

	// Convert the response to our internal type (same logic as Create)
	result := &SecurityDetectionRuleResponse{
		ID:          ruleResponse.Id.String(),
		Name:        ruleResponse.Name,
		Description: string(ruleResponse.Description),
		Type:        string(ruleResponse.Type),
		Severity:    string(ruleResponse.Severity),
		Risk:        ruleResponse.RiskScore,
		Enabled:     bool(ruleResponse.Enabled),
		From:        string(ruleResponse.From),
		To:          string(ruleResponse.To),
		Interval:    string(ruleResponse.Interval),
		Version:     ruleResponse.Version,
		MaxSignals:  ruleResponse.MaxSignals,
		CreatedAt:   ruleResponse.CreatedAt.Format("2006-01-02T15:04:05.000Z"),
		CreatedBy:   ruleResponse.CreatedBy,
		UpdatedAt:   ruleResponse.UpdatedAt.Format("2006-01-02T15:04:05.000Z"),
		UpdatedBy:   ruleResponse.UpdatedBy,
	}

	// Set optional fields (same logic as Create)
	queryStr := string(ruleResponse.Query)
	result.Query = &queryStr

	langStr := string(ruleResponse.Language)
	result.Language = &langStr

	if ruleResponse.Index != nil {
		result.Index = make([]string, len(*ruleResponse.Index))
		for i, idx := range *ruleResponse.Index {
			result.Index[i] = idx
		}
	}
	result.Tags = make([]string, len(ruleResponse.Tags))
	for i, tag := range ruleResponse.Tags {
		result.Tags[i] = tag
	}
	result.Author = make([]string, len(ruleResponse.Author))
	for i, author := range ruleResponse.Author {
		result.Author[i] = author
	}
	if ruleResponse.License != nil {
		licenseStr := string(*ruleResponse.License)
		result.License = &licenseStr
	}
	if ruleResponse.RuleNameOverride != nil {
		overrideStr := string(*ruleResponse.RuleNameOverride)
		result.RuleNameOverride = &overrideStr
	}
	if ruleResponse.TimestampOverride != nil {
		timestampStr := string(*ruleResponse.TimestampOverride)
		result.TimestampOverride = &timestampStr
	}
	if ruleResponse.Note != nil {
		noteStr := string(*ruleResponse.Note)
		result.Note = &noteStr
	}
	result.References = make([]string, len(ruleResponse.References))
	for i, ref := range ruleResponse.References {
		result.References[i] = ref
	}
	result.FalsePositives = make([]string, len(ruleResponse.FalsePositives))
	for i, fp := range ruleResponse.FalsePositives {
		result.FalsePositives[i] = fp
	}

	return result, diags
}

// UpdateSecurityDetectionRule updates an existing security detection rule using the generated API client
func UpdateSecurityDetectionRule(ctx context.Context, client *clients.ApiClient, spaceId, ruleId string, rule *SecurityDetectionRuleRequest) (*SecurityDetectionRuleResponse, diag.Diagnostics) {
	var diags diag.Diagnostics

	kbClient, err := client.GetKibanaClient()
	if err != nil {
		diags.AddError("Failed to get Kibana client", err.Error())
		return nil, diags
	}

	// Create the generated API client
	genClient, err := kbapi.NewClientWithResponses(kbClient.Client.BaseURL, kbapi.WithHTTPClient(kbClient.Client.GetClient()))
	if err != nil {
		diags.AddError("Failed to create generated client", err.Error())
		return nil, diags
	}

	// Convert our request to the generated API types for update
	updateProps := kbapi.SecurityDetectionsAPIRuleUpdateProps{}
	
	// Create a QueryRuleUpdateProps
	parsedId, err := uuid.Parse(ruleId)
	if err != nil {
		diags.AddError("Invalid rule ID", fmt.Sprintf("Failed to parse rule ID as UUID: %s", err.Error()))
		return nil, diags
	}
	id := kbapi.SecurityDetectionsAPIRuleObjectId(parsedId)
	enabled := kbapi.SecurityDetectionsAPIIsRuleEnabled(rule.Enabled)
	from := kbapi.SecurityDetectionsAPIRuleIntervalFrom(rule.From)
	to := kbapi.SecurityDetectionsAPIRuleIntervalTo(rule.To)
	interval := kbapi.SecurityDetectionsAPIRuleInterval(rule.Interval)

	queryRuleProps := kbapi.SecurityDetectionsAPIQueryRuleUpdateProps{
		Id:          &id,
		Name:        rule.Name,
		Description: kbapi.SecurityDetectionsAPIRuleDescription(rule.Description),
		Type:        kbapi.SecurityDetectionsAPIQueryRuleUpdatePropsType(rule.Type),
		Severity:    kbapi.SecurityDetectionsAPISeverity(rule.Severity),
		RiskScore:   rule.Risk,
		Enabled:     &enabled,
		From:        &from,
		To:          &to,
		Interval:    &interval,
		MaxSignals:  &rule.MaxSignals,
		Version:     &rule.Version,
	}

	// Set optional fields (same logic as Create)
	if rule.Query != nil {
		query := kbapi.SecurityDetectionsAPIRuleQuery(*rule.Query)
		queryRuleProps.Query = &query
	}
	if rule.Language != nil {
		language := kbapi.SecurityDetectionsAPIKqlQueryLanguage(*rule.Language)
		queryRuleProps.Language = &language
	}
	if len(rule.Index) > 0 {
		indexArray := make(kbapi.SecurityDetectionsAPIIndexPatternArray, len(rule.Index))
		for i, idx := range rule.Index {
			indexArray[i] = idx
		}
		queryRuleProps.Index = &indexArray
	}
	if len(rule.Tags) > 0 {
		tagArray := make(kbapi.SecurityDetectionsAPIRuleTagArray, len(rule.Tags))
		for i, tag := range rule.Tags {
			tagArray[i] = tag
		}
		queryRuleProps.Tags = &tagArray
	}
	if len(rule.Author) > 0 {
		authorArray := make(kbapi.SecurityDetectionsAPIRuleAuthorArray, len(rule.Author))
		for i, author := range rule.Author {
			authorArray[i] = author
		}
		queryRuleProps.Author = &authorArray
	}
	if rule.License != nil {
		license := kbapi.SecurityDetectionsAPIRuleLicense(*rule.License)
		queryRuleProps.License = &license
	}
	if rule.RuleNameOverride != nil {
		override := kbapi.SecurityDetectionsAPIRuleNameOverride(*rule.RuleNameOverride)
		queryRuleProps.RuleNameOverride = &override
	}
	if rule.TimestampOverride != nil {
		timestampOverride := kbapi.SecurityDetectionsAPITimestampOverride(*rule.TimestampOverride)
		queryRuleProps.TimestampOverride = &timestampOverride
	}
	if rule.Note != nil {
		note := kbapi.SecurityDetectionsAPIInvestigationGuide(*rule.Note)
		queryRuleProps.Note = &note
	}
	if len(rule.References) > 0 {
		refArray := make(kbapi.SecurityDetectionsAPIRuleReferenceArray, len(rule.References))
		for i, ref := range rule.References {
			refArray[i] = ref
		}
		queryRuleProps.References = &refArray
	}
	if len(rule.FalsePositives) > 0 {
		fpArray := make(kbapi.SecurityDetectionsAPIRuleFalsePositiveArray, len(rule.FalsePositives))
		for i, fp := range rule.FalsePositives {
			fpArray[i] = fp
		}
		queryRuleProps.FalsePositives = &fpArray
	}

	// Set the query rule props in the union
	if err := updateProps.FromSecurityDetectionsAPIQueryRuleUpdateProps(queryRuleProps); err != nil {
		diags.AddError("Failed to create request body", err.Error())
		return nil, diags
	}

	// Call the generated API
	resp, err := genClient.UpdateRuleWithResponse(ctx, kbapi.SpaceId(spaceId), updateProps)
	if err != nil {
		diags.AddError("Failed to execute request", err.Error())
		return nil, diags
	}

	// Check for API errors
	if resp.StatusCode() >= 300 {
		diags.AddError(
			"API request failed",
			fmt.Sprintf("Status: %d, Body: %s", resp.StatusCode(), string(resp.Body)),
		)
		return nil, diags
	}

	// Parse the response
	if resp.JSON200 == nil {
		diags.AddError("Unexpected response", "Expected JSON response but got nil")
		return nil, diags
	}

	ruleResponse, err := resp.JSON200.AsSecurityDetectionsAPIQueryRule()
	if err != nil {
		diags.AddError("Failed to parse response", err.Error())
		return nil, diags
	}

	// Convert the response to our internal type (same logic as Create/Read)
	result := &SecurityDetectionRuleResponse{
		ID:          ruleResponse.Id.String(),
		Name:        ruleResponse.Name,
		Description: string(ruleResponse.Description),
		Type:        string(ruleResponse.Type),
		Severity:    string(ruleResponse.Severity),
		Risk:        ruleResponse.RiskScore,
		Enabled:     bool(ruleResponse.Enabled),
		From:        string(ruleResponse.From),
		To:          string(ruleResponse.To),
		Interval:    string(ruleResponse.Interval),
		Version:     ruleResponse.Version,
		MaxSignals:  ruleResponse.MaxSignals,
		CreatedAt:   ruleResponse.CreatedAt.Format("2006-01-02T15:04:05.000Z"),
		CreatedBy:   ruleResponse.CreatedBy,
		UpdatedAt:   ruleResponse.UpdatedAt.Format("2006-01-02T15:04:05.000Z"),
		UpdatedBy:   ruleResponse.UpdatedBy,
	}

	// Set optional fields (same logic as Create/Read)
	queryStr := string(ruleResponse.Query)
	result.Query = &queryStr

	langStr := string(ruleResponse.Language)
	result.Language = &langStr

	if ruleResponse.Index != nil {
		result.Index = make([]string, len(*ruleResponse.Index))
		for i, idx := range *ruleResponse.Index {
			result.Index[i] = idx
		}
	}
	result.Tags = make([]string, len(ruleResponse.Tags))
	for i, tag := range ruleResponse.Tags {
		result.Tags[i] = tag
	}
	result.Author = make([]string, len(ruleResponse.Author))
	for i, author := range ruleResponse.Author {
		result.Author[i] = author
	}
	if ruleResponse.License != nil {
		licenseStr := string(*ruleResponse.License)
		result.License = &licenseStr
	}
	if ruleResponse.RuleNameOverride != nil {
		overrideStr := string(*ruleResponse.RuleNameOverride)
		result.RuleNameOverride = &overrideStr
	}
	if ruleResponse.TimestampOverride != nil {
		timestampStr := string(*ruleResponse.TimestampOverride)
		result.TimestampOverride = &timestampStr
	}
	if ruleResponse.Note != nil {
		noteStr := string(*ruleResponse.Note)
		result.Note = &noteStr
	}
	result.References = make([]string, len(ruleResponse.References))
	for i, ref := range ruleResponse.References {
		result.References[i] = ref
	}
	result.FalsePositives = make([]string, len(ruleResponse.FalsePositives))
	for i, fp := range ruleResponse.FalsePositives {
		result.FalsePositives[i] = fp
	}

	return result, diags
}

// DeleteSecurityDetectionRule deletes a security detection rule by ID using the generated API client
func DeleteSecurityDetectionRule(ctx context.Context, client *clients.ApiClient, spaceId, ruleId string) diag.Diagnostics {
	var diags diag.Diagnostics

	kbClient, err := client.GetKibanaClient()
	if err != nil {
		diags.AddError("Failed to get Kibana client", err.Error())
		return diags
	}

	// Create the generated API client
	genClient, err := kbapi.NewClientWithResponses(kbClient.Client.BaseURL, kbapi.WithHTTPClient(kbClient.Client.GetClient()))
	if err != nil {
		diags.AddError("Failed to create generated client", err.Error())
		return diags
	}

	// Set up parameters - use rule ID for deletion
	parsedId, err := uuid.Parse(ruleId)
	if err != nil {
		diags.AddError("Invalid rule ID", fmt.Sprintf("Failed to parse rule ID as UUID: %s", err.Error()))
		return diags
	}
	id := kbapi.SecurityDetectionsAPIRuleObjectId(parsedId)
	params := &kbapi.DeleteRuleParams{
		Id: &id,
	}

	// Call the generated API
	resp, err := genClient.DeleteRuleWithResponse(ctx, kbapi.SpaceId(spaceId), params)
	if err != nil {
		diags.AddError("Failed to execute request", err.Error())
		return diags
	}

	// Handle not found (rule might already be deleted)
	if resp.StatusCode() == 404 {
		return diags // Already deleted, no error
	}

	// Check for other API errors
	if resp.StatusCode() >= 300 {
		diags.AddError(
			"API request failed",
			fmt.Sprintf("Status: %d, Body: %s", resp.StatusCode(), string(resp.Body)),
		)
		return diags
	}

	return diags
}
