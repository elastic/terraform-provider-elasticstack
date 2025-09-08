package security_detection_rule

import (
	"context"
	"fmt"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *securityDetectionRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SecurityDetectionRuleData

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Parse ID to get space_id and rule_id
	spaceId, ruleId, diags := r.parseResourceId(data.Id.ValueString())
	resp.Diagnostics.Append(diags...)
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

	// Read the rule
	ruleObjectId := kbapi.SecurityDetectionsAPIRuleObjectId(uuid.MustParse(ruleId))
	params := &kbapi.ReadRuleParams{
		Id: &ruleObjectId,
	}

	response, err := kbClient.API.ReadRuleWithResponse(ctx, params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading security detection rule",
			"Could not read security detection rule: "+err.Error(),
		)
		return
	}

	if response.StatusCode() == 404 {
		// Rule was deleted outside of Terraform
		resp.State.RemoveResource(ctx)
		return
	}

	if response.StatusCode() != 200 {
		resp.Diagnostics.AddError(
			"Error reading security detection rule",
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

	// Ensure space_id is set correctly
	data.SpaceId = types.StringValue(spaceId)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *securityDetectionRuleResource) parseResourceId(id string) (spaceId, ruleId string, diags diag.Diagnostics) {
	parts := strings.Split(id, "/")
	if len(parts) != 2 {
		diags.AddError(
			"Invalid resource ID format",
			fmt.Sprintf("Expected format 'space_id/rule_id', got: %s", id),
		)
		return
	}
	return parts[0], parts[1], diags
}

func (r *securityDetectionRuleResource) parseRuleResponse(ctx context.Context, response *kbapi.SecurityDetectionsAPIRuleResponse) (*kbapi.SecurityDetectionsAPIQueryRule, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Since we only support query rules for now, try to parse as query rule
	queryRule, err := response.AsSecurityDetectionsAPIQueryRule()
	if err != nil {
		diags.AddError(
			"Error parsing rule response",
			"Could not parse rule as query rule: "+err.Error(),
		)
		return nil, diags
	}

	return &queryRule, diags
}

func (r *securityDetectionRuleResource) updateDataFromRule(ctx context.Context, data *SecurityDetectionRuleData, rule *kbapi.SecurityDetectionsAPIQueryRule) diag.Diagnostics {
	var diags diag.Diagnostics

	// Update core fields
	data.RuleId = types.StringValue(string(rule.RuleId))
	data.Name = types.StringValue(string(rule.Name))
	data.Type = types.StringValue(string(rule.Type))
	data.Query = types.StringValue(rule.Query)
	data.Language = types.StringValue(string(rule.Language))
	data.Enabled = types.BoolValue(bool(rule.Enabled))
	data.From = types.StringValue(string(rule.From))
	data.To = types.StringValue(string(rule.To))
	data.Interval = types.StringValue(string(rule.Interval))
	data.Description = types.StringValue(string(rule.Description))
	data.RiskScore = types.Int64Value(int64(rule.RiskScore))
	data.Severity = types.StringValue(string(rule.Severity))
	data.MaxSignals = types.Int64Value(int64(rule.MaxSignals))
	data.Version = types.Int64Value(int64(rule.Version))

	// Update read-only fields
	data.CreatedAt = types.StringValue(rule.CreatedAt.Format("2006-01-02T15:04:05.000Z"))
	data.CreatedBy = types.StringValue(rule.CreatedBy)
	data.UpdatedAt = types.StringValue(rule.UpdatedAt.Format("2006-01-02T15:04:05.000Z"))
	data.UpdatedBy = types.StringValue(rule.UpdatedBy)
	data.Revision = types.Int64Value(int64(rule.Revision))

	// Update index patterns
	if rule.Index != nil && len(*rule.Index) > 0 {
		indexList := make([]string, len(*rule.Index))
		for i, idx := range *rule.Index {
			indexList[i] = string(idx)
		}
		indexListValue, diagsIndex := types.ListValueFrom(ctx, types.StringType, indexList)
		diags.Append(diagsIndex...)
		data.Index = indexListValue
	} else {
		data.Index = types.ListValueMust(types.StringType, []attr.Value{})
	}

	// Update author
	if len(rule.Author) > 0 {
		authorList := make([]string, len(rule.Author))
		//nolint:staticcheck // Type conversion required, can't use copy()
		for i, author := range rule.Author {
			authorList[i] = author
		}
		authorListValue, diagsAuthor := types.ListValueFrom(ctx, types.StringType, authorList)
		diags.Append(diagsAuthor...)
		data.Author = authorListValue
	} else {
		data.Author = types.ListValueMust(types.StringType, []attr.Value{})
	}

	// Update tags
	if len(rule.Tags) > 0 {
		tagsList := make([]string, len(rule.Tags))
		//nolint:staticcheck // Type conversion required, can't use copy()
		for i, tag := range rule.Tags {
			tagsList[i] = tag
		}
		tagsListValue, diagsTags := types.ListValueFrom(ctx, types.StringType, tagsList)
		diags.Append(diagsTags...)
		data.Tags = tagsListValue
	} else {
		data.Tags = types.ListValueMust(types.StringType, []attr.Value{})
	}

	// Update false positives
	if len(rule.FalsePositives) > 0 {
		fpList := make([]string, len(rule.FalsePositives))
		//nolint:staticcheck // Type conversion required, can't use copy()
		for i, fp := range rule.FalsePositives {
			fpList[i] = fp
		}
		fpListValue, diagsFp := types.ListValueFrom(ctx, types.StringType, fpList)
		diags.Append(diagsFp...)
		data.FalsePositives = fpListValue
	} else {
		data.FalsePositives = types.ListValueMust(types.StringType, []attr.Value{})
	}

	// Update references
	if len(rule.References) > 0 {
		refList := make([]string, len(rule.References))
		for i, ref := range rule.References {
			refList[i] = string(ref)
		}
		refListValue, diagsRef := types.ListValueFrom(ctx, types.StringType, refList)
		diags.Append(diagsRef...)
		data.References = refListValue
	} else {
		data.References = types.ListValueMust(types.StringType, []attr.Value{})
	}

	// Update optional string fields
	if rule.License != nil {
		data.License = types.StringValue(string(*rule.License))
	} else {
		data.License = types.StringNull()
	}

	if rule.Note != nil {
		data.Note = types.StringValue(string(*rule.Note))
	} else {
		data.Note = types.StringNull()
	}

	data.Setup = types.StringValue(string(rule.Setup))

	return diags
}
