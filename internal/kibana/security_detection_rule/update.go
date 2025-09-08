package security_detection_rule

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *securityDetectionRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data SecurityDetectionRuleData

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the rule using kbapi client
	kbClient, err := r.client.GetKibanaOapiClient()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error getting Kibana client",
			"Could not get Kibana OAPI client: "+err.Error(),
		)
		return
	}

	// Build the update request
	updateProps, diags := r.buildUpdateProps(ctx, data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update the rule
	response, err := kbClient.API.UpdateRuleWithResponse(ctx, updateProps)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating security detection rule",
			"Could not update security detection rule: "+err.Error(),
		)
		return
	}

	if response.StatusCode() != 200 {
		resp.Diagnostics.AddError(
			"Error updating security detection rule",
			fmt.Sprintf("API returned status %d: %s", response.StatusCode(), string(response.Body)),
		)
		return
	}

	// Parse the response
	ruleResponse, diags := r.parseRuleResponse(ctx, response.JSON200)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update the data with response values
	diags = r.updateDataFromRule(ctx, &data, ruleResponse)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *securityDetectionRuleResource) buildUpdateProps(ctx context.Context, data SecurityDetectionRuleData) (kbapi.SecurityDetectionsAPIRuleUpdateProps, diag.Diagnostics) {
	var diags diag.Diagnostics
	var updateProps kbapi.SecurityDetectionsAPIRuleUpdateProps

	queryRuleQuery := kbapi.SecurityDetectionsAPIRuleQuery(data.Query.ValueString())
	ruleObjectId := kbapi.SecurityDetectionsAPIRuleObjectId(uuid.MustParse(data.RuleId.ValueString()))
	// Convert data to QueryRuleUpdateProps since we're only supporting query rules initially
	queryRule := kbapi.SecurityDetectionsAPIQueryRuleUpdateProps{
		Id:          &ruleObjectId,
		Name:        kbapi.SecurityDetectionsAPIRuleName(data.Name.ValueString()),
		Description: kbapi.SecurityDetectionsAPIRuleDescription(data.Description.ValueString()),
		Type:        kbapi.SecurityDetectionsAPIQueryRuleUpdatePropsType("query"),
		Query:       &queryRuleQuery,
		RiskScore:   kbapi.SecurityDetectionsAPIRiskScore(data.RiskScore.ValueInt64()),
		Severity:    kbapi.SecurityDetectionsAPISeverity(data.Severity.ValueString()),
	}

	// For updates, we need to include the rule_id
	ruleId := kbapi.SecurityDetectionsAPIRuleSignatureId(data.RuleId.ValueString())
	queryRule.RuleId = &ruleId

	// Set enabled status
	if !data.Enabled.IsNull() {
		enabled := kbapi.SecurityDetectionsAPIIsRuleEnabled(data.Enabled.ValueBool())
		queryRule.Enabled = &enabled
	}

	// Set query language
	if !data.Language.IsNull() {
		var language kbapi.SecurityDetectionsAPIKqlQueryLanguage
		switch data.Language.ValueString() {
		case "kuery":
			language = "kuery"
		case "lucene":
			language = "lucene"
		default:
			language = "kuery"
		}
		queryRule.Language = &language
	}

	// Set time range
	if !data.From.IsNull() {
		from := kbapi.SecurityDetectionsAPIRuleIntervalFrom(data.From.ValueString())
		queryRule.From = &from
	}

	if !data.To.IsNull() {
		to := kbapi.SecurityDetectionsAPIRuleIntervalTo(data.To.ValueString())
		queryRule.To = &to
	}

	// Set interval
	if !data.Interval.IsNull() {
		interval := kbapi.SecurityDetectionsAPIRuleInterval(data.Interval.ValueString())
		queryRule.Interval = &interval
	}

	// Set index patterns
	if !data.Index.IsNull() && !data.Index.IsUnknown() {
		var indexList []string
		diags.Append(data.Index.ElementsAs(ctx, &indexList, false)...)
		if !diags.HasError() {
			indexPatterns := make(kbapi.SecurityDetectionsAPIIndexPatternArray, len(indexList))
			//nolint:staticcheck // Type conversion required, can't use copy()
			for i, idx := range indexList {
				indexPatterns[i] = idx
			}
			queryRule.Index = &indexPatterns
		}
	}

	// Set author
	if !data.Author.IsNull() && !data.Author.IsUnknown() {
		var authorList []string
		diags.Append(data.Author.ElementsAs(ctx, &authorList, false)...)
		if !diags.HasError() {
			authorArray := make(kbapi.SecurityDetectionsAPIRuleAuthorArray, len(authorList))
			//nolint:staticcheck // Type conversion required, can't use copy()
			for i, author := range authorList {
				authorArray[i] = author
			}
			queryRule.Author = &authorArray
		}
	}

	// Set tags
	if !data.Tags.IsNull() && !data.Tags.IsUnknown() {
		var tagsList []string
		diags.Append(data.Tags.ElementsAs(ctx, &tagsList, false)...)
		if !diags.HasError() {
			tagsArray := make(kbapi.SecurityDetectionsAPIRuleTagArray, len(tagsList))
			//nolint:staticcheck // Type conversion required, can't use copy()
			for i, tag := range tagsList {
				tagsArray[i] = tag
			}
			queryRule.Tags = &tagsArray
		}
	}

	// Set false positives
	if !data.FalsePositives.IsNull() && !data.FalsePositives.IsUnknown() {
		var fpList []string
		diags.Append(data.FalsePositives.ElementsAs(ctx, &fpList, false)...)
		if !diags.HasError() {
			fpArray := make(kbapi.SecurityDetectionsAPIRuleFalsePositiveArray, len(fpList))
			//nolint:staticcheck // Type conversion required, can't use copy()
			for i, fp := range fpList {
				fpArray[i] = fp
			}
			queryRule.FalsePositives = &fpArray
		}
	}

	// Set references
	if !data.References.IsNull() && !data.References.IsUnknown() {
		var refList []string
		diags.Append(data.References.ElementsAs(ctx, &refList, false)...)
		if !diags.HasError() {
			refArray := make(kbapi.SecurityDetectionsAPIRuleReferenceArray, len(refList))
			//nolint:staticcheck // Type conversion required, can't use copy()
			for i, ref := range refList {
				refArray[i] = ref
			}
			queryRule.References = &refArray
		}
	}

	// Set optional string fields
	if !data.License.IsNull() {
		license := kbapi.SecurityDetectionsAPIRuleLicense(data.License.ValueString())
		queryRule.License = &license
	}

	if !data.Note.IsNull() {
		note := kbapi.SecurityDetectionsAPIInvestigationGuide(data.Note.ValueString())
		queryRule.Note = &note
	}

	if !data.Setup.IsNull() {
		setup := kbapi.SecurityDetectionsAPISetupGuide(data.Setup.ValueString())
		queryRule.Setup = &setup
	} else {
		// Set to empty string if not provided (required field in update)
		setup := kbapi.SecurityDetectionsAPISetupGuide("")
		queryRule.Setup = &setup
	}

	// Set max signals
	if !data.MaxSignals.IsNull() {
		maxSignals := kbapi.SecurityDetectionsAPIMaxSignals(data.MaxSignals.ValueInt64())
		queryRule.MaxSignals = &maxSignals
	}

	// Set version
	if !data.Version.IsNull() {
		version := kbapi.SecurityDetectionsAPIRuleVersion(data.Version.ValueInt64())
		queryRule.Version = &version
	}

	// Convert to union type
	err := updateProps.FromSecurityDetectionsAPIQueryRuleUpdateProps(queryRule)
	if err != nil {
		diags.AddError(
			"Error building update properties",
			"Could not convert rule properties: "+err.Error(),
		)
	}

	return updateProps, diags
}
