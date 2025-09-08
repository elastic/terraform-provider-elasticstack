package detection_rule

import (
	"context"
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *securityDetectionRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data SecurityDetectionRuleData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert the data to API request
	apiRequest, diags := dataToAPIRequest(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get space ID
	spaceId := data.SpaceId.ValueString()
	if spaceId == "" {
		spaceId = "default"
	}

	// Create the rule
	var ruleId *string
	if !data.RuleId.IsNull() && !data.RuleId.IsUnknown() {
		id := data.RuleId.ValueString()
		ruleId = &id
	}

	result, diags := CreateSecurityDetectionRule(ctx, r.client, spaceId, ruleId, apiRequest)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update the data with the response
	diags = apiResponseToData(ctx, result, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create composite ID for state management
	compositeID := &clients.CompositeId{
		ClusterId:  spaceId,
		ResourceId: result.ID,
	}
	data.Id = types.StringValue(compositeID.String())

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func dataToAPIRequest(ctx context.Context, data *SecurityDetectionRuleData) (*SecurityDetectionRuleRequest, diag.Diagnostics) {
	var diags diag.Diagnostics

	req := &SecurityDetectionRuleRequest{
		Name:        data.Name.ValueString(),
		Description: data.Description.ValueString(),
		Type:        data.Type.ValueString(),
		Severity:    data.Severity.ValueString(),
		Risk:        int(data.Risk.ValueInt64()),
		Enabled:     data.Enabled.ValueBool(),
		From:        data.From.ValueString(),
		To:          data.To.ValueString(),
		Interval:    data.Interval.ValueString(),
		Version:     int(data.Version.ValueInt64()),
		MaxSignals:  int(data.MaxSignals.ValueInt64()),
	}

	// Handle optional string fields
	if !data.Query.IsNull() && !data.Query.IsUnknown() {
		query := data.Query.ValueString()
		req.Query = &query
	}

	if !data.Language.IsNull() && !data.Language.IsUnknown() {
		language := data.Language.ValueString()
		req.Language = &language
	}

	if !data.License.IsNull() && !data.License.IsUnknown() {
		license := data.License.ValueString()
		req.License = &license
	}

	if !data.RuleNameOverride.IsNull() && !data.RuleNameOverride.IsUnknown() {
		override := data.RuleNameOverride.ValueString()
		req.RuleNameOverride = &override
	}

	if !data.TimestampOverride.IsNull() && !data.TimestampOverride.IsUnknown() {
		override := data.TimestampOverride.ValueString()
		req.TimestampOverride = &override
	}

	if !data.Note.IsNull() && !data.Note.IsUnknown() {
		note := data.Note.ValueString()
		req.Note = &note
	}

	// Handle Meta as JSON
	if !data.Meta.IsNull() && !data.Meta.IsUnknown() {
		var meta map[string]any
		err := json.Unmarshal([]byte(data.Meta.ValueString()), &meta)
		if err != nil {
			diags.AddError("Invalid meta JSON", err.Error())
			return nil, diags
		}
		req.Meta = &meta
	}

	// Handle string arrays
	if !data.Index.IsNull() && !data.Index.IsUnknown() {
		var indices []string
		diags.Append(data.Index.ElementsAs(ctx, &indices, false)...)
		if diags.HasError() {
			return nil, diags
		}
		req.Index = indices
	}

	if !data.Tags.IsNull() && !data.Tags.IsUnknown() {
		var tags []string
		diags.Append(data.Tags.ElementsAs(ctx, &tags, false)...)
		if diags.HasError() {
			return nil, diags
		}
		req.Tags = tags
	}

	if !data.Author.IsNull() && !data.Author.IsUnknown() {
		var authors []string
		diags.Append(data.Author.ElementsAs(ctx, &authors, false)...)
		if diags.HasError() {
			return nil, diags
		}
		req.Author = authors
	}

	if !data.References.IsNull() && !data.References.IsUnknown() {
		var references []string
		diags.Append(data.References.ElementsAs(ctx, &references, false)...)
		if diags.HasError() {
			return nil, diags
		}
		req.References = references
	}

	if !data.FalsePositives.IsNull() && !data.FalsePositives.IsUnknown() {
		var falsePositives []string
		diags.Append(data.FalsePositives.ElementsAs(ctx, &falsePositives, false)...)
		if diags.HasError() {
			return nil, diags
		}
		req.FalsePositives = falsePositives
	}

	// Handle exceptions list (for now, just as string array)
	if !data.ExceptionsList.IsNull() && !data.ExceptionsList.IsUnknown() {
		var exceptions []string
		diags.Append(data.ExceptionsList.ElementsAs(ctx, &exceptions, false)...)
		if diags.HasError() {
			return nil, diags
		}
		// Convert to []any for JSON serialization
		req.ExceptionsList = make([]any, len(exceptions))
		for i, ex := range exceptions {
			req.ExceptionsList[i] = ex
		}
	}

	return req, diags
}

func apiResponseToData(ctx context.Context, result *SecurityDetectionRuleResponse, data *SecurityDetectionRuleData) diag.Diagnostics {
	var diags diag.Diagnostics

	data.RuleId = types.StringValue(result.ID)
	data.Name = types.StringValue(result.Name)
	data.Description = types.StringValue(result.Description)
	data.Type = types.StringValue(result.Type)
	data.Severity = types.StringValue(result.Severity)
	data.Risk = types.Int64Value(int64(result.Risk))
	data.Enabled = types.BoolValue(result.Enabled)
	data.From = types.StringValue(result.From)
	data.To = types.StringValue(result.To)
	data.Interval = types.StringValue(result.Interval)
	data.Version = types.Int64Value(int64(result.Version))
	data.MaxSignals = types.Int64Value(int64(result.MaxSignals))

	// Handle optional fields
	if result.Query != nil {
		data.Query = types.StringValue(*result.Query)
	} else {
		data.Query = types.StringNull()
	}

	if result.Language != nil {
		data.Language = types.StringValue(*result.Language)
	} else {
		data.Language = types.StringValue("kuery") // Default value
	}

	if result.License != nil {
		data.License = types.StringValue(*result.License)
	} else {
		data.License = types.StringNull()
	}

	if result.RuleNameOverride != nil {
		data.RuleNameOverride = types.StringValue(*result.RuleNameOverride)
	} else {
		data.RuleNameOverride = types.StringNull()
	}

	if result.TimestampOverride != nil {
		data.TimestampOverride = types.StringValue(*result.TimestampOverride)
	} else {
		data.TimestampOverride = types.StringNull()
	}

	if result.Note != nil {
		data.Note = types.StringValue(*result.Note)
	} else {
		data.Note = types.StringNull()
	}

	// Handle Meta as JSON string
	if result.Meta != nil {
		metaBytes, err := json.Marshal(result.Meta)
		if err != nil {
			diags.AddError("Failed to marshal meta", err.Error())
			return diags
		}
		data.Meta = types.StringValue(string(metaBytes))
	} else {
		data.Meta = types.StringNull()
	}

	// Handle arrays
	if len(result.Index) > 0 {
		indexValues := make([]types.String, len(result.Index))
		for i, idx := range result.Index {
			indexValues[i] = types.StringValue(idx)
		}
		data.Index, _ = types.ListValueFrom(ctx, types.StringType, indexValues)
	} else {
		// Default to wildcard index
		data.Index, _ = types.ListValueFrom(ctx, types.StringType, []types.String{types.StringValue("*")})
	}

	if len(result.Tags) > 0 {
		tagValues := make([]types.String, len(result.Tags))
		for i, tag := range result.Tags {
			tagValues[i] = types.StringValue(tag)
		}
		data.Tags, _ = types.ListValueFrom(ctx, types.StringType, tagValues)
	} else {
		data.Tags, _ = types.ListValueFrom(ctx, types.StringType, []types.String{})
	}

	if len(result.Author) > 0 {
		authorValues := make([]types.String, len(result.Author))
		for i, author := range result.Author {
			authorValues[i] = types.StringValue(author)
		}
		data.Author, _ = types.ListValueFrom(ctx, types.StringType, authorValues)
	} else {
		data.Author, _ = types.ListValueFrom(ctx, types.StringType, []types.String{})
	}

	if len(result.References) > 0 {
		refValues := make([]types.String, len(result.References))
		for i, ref := range result.References {
			refValues[i] = types.StringValue(ref)
		}
		data.References, _ = types.ListValueFrom(ctx, types.StringType, refValues)
	} else {
		data.References, _ = types.ListValueFrom(ctx, types.StringType, []types.String{})
	}

	if len(result.FalsePositives) > 0 {
		fpValues := make([]types.String, len(result.FalsePositives))
		for i, fp := range result.FalsePositives {
			fpValues[i] = types.StringValue(fp)
		}
		data.FalsePositives, _ = types.ListValueFrom(ctx, types.StringType, fpValues)
	} else {
		data.FalsePositives, _ = types.ListValueFrom(ctx, types.StringType, []types.String{})
	}

	if len(result.ExceptionsList) > 0 {
		// Convert exceptions to strings (simplified)
		excValues := make([]types.String, len(result.ExceptionsList))
		for i, exc := range result.ExceptionsList {
			if excStr, ok := exc.(string); ok {
				excValues[i] = types.StringValue(excStr)
			} else {
				// Convert complex exceptions to JSON strings
				excBytes, _ := json.Marshal(exc)
				excValues[i] = types.StringValue(string(excBytes))
			}
		}
		data.ExceptionsList, _ = types.ListValueFrom(ctx, types.StringType, excValues)
	} else {
		data.ExceptionsList, _ = types.ListValueFrom(ctx, types.StringType, []types.String{})
	}

	return diags
}
