package security_detection_rule

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type SecurityDetectionRuleData struct {
	Id       types.String `tfsdk:"id"`
	SpaceId  types.String `tfsdk:"space_id"`
	RuleId   types.String `tfsdk:"rule_id"`
	Name     types.String `tfsdk:"name"`
	Type     types.String `tfsdk:"type"`
	Query    types.String `tfsdk:"query"`
	Language types.String `tfsdk:"language"`
	Index    types.List   `tfsdk:"index"`
	Enabled  types.Bool   `tfsdk:"enabled"`
	From     types.String `tfsdk:"from"`
	To       types.String `tfsdk:"to"`
	Interval types.String `tfsdk:"interval"`

	// Rule content
	Description types.String `tfsdk:"description"`
	RiskScore   types.Int64  `tfsdk:"risk_score"`
	Severity    types.String `tfsdk:"severity"`
	Author      types.List   `tfsdk:"author"`
	Tags        types.List   `tfsdk:"tags"`
	License     types.String `tfsdk:"license"`

	// Optional fields
	FalsePositives types.List   `tfsdk:"false_positives"`
	References     types.List   `tfsdk:"references"`
	Note           types.String `tfsdk:"note"`
	Setup          types.String `tfsdk:"setup"`
	MaxSignals     types.Int64  `tfsdk:"max_signals"`
	Version        types.Int64  `tfsdk:"version"`

	// Read-only fields
	CreatedAt types.String `tfsdk:"created_at"`
	CreatedBy types.String `tfsdk:"created_by"`
	UpdatedAt types.String `tfsdk:"updated_at"`
	UpdatedBy types.String `tfsdk:"updated_by"`
	Revision  types.Int64  `tfsdk:"revision"`
}

func (d SecurityDetectionRuleData) toCreateProps(ctx context.Context) (kbapi.SecurityDetectionsAPIRuleCreateProps, diag.Diagnostics) {
	var diags diag.Diagnostics
	var createProps kbapi.SecurityDetectionsAPIRuleCreateProps

	queryRuleQuery := kbapi.SecurityDetectionsAPIRuleQuery(d.Query.ValueString())
	// Convert data to QueryRuleCreateProps since we're only supporting query rules initially
	queryRule := kbapi.SecurityDetectionsAPIQueryRuleCreateProps{
		Name:        kbapi.SecurityDetectionsAPIRuleName(d.Name.ValueString()),
		Description: kbapi.SecurityDetectionsAPIRuleDescription(d.Description.ValueString()),
		Type:        kbapi.SecurityDetectionsAPIQueryRuleCreatePropsType("query"),
		Query:       &queryRuleQuery,
		RiskScore:   kbapi.SecurityDetectionsAPIRiskScore(d.RiskScore.ValueInt64()),
		Severity:    kbapi.SecurityDetectionsAPISeverity(d.Severity.ValueString()),
	}

	// Set optional rule_id if provided
	if utils.IsKnown(d.RuleId) {
		ruleId := kbapi.SecurityDetectionsAPIRuleSignatureId(d.RuleId.ValueString())
		queryRule.RuleId = &ruleId
	}

	// Set enabled status
	if utils.IsKnown(d.Enabled) {
		enabled := kbapi.SecurityDetectionsAPIIsRuleEnabled(d.Enabled.ValueBool())
		queryRule.Enabled = &enabled
	}

	// Set query language
	if utils.IsKnown(d.Language) {
		var language kbapi.SecurityDetectionsAPIKqlQueryLanguage
		switch d.Language.ValueString() {
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
	if utils.IsKnown(d.From) {
		from := kbapi.SecurityDetectionsAPIRuleIntervalFrom(d.From.ValueString())
		queryRule.From = &from
	}

	if utils.IsKnown(d.To) {
		to := kbapi.SecurityDetectionsAPIRuleIntervalTo(d.To.ValueString())
		queryRule.To = &to
	}

	// Set interval
	if utils.IsKnown(d.Interval) {
		interval := kbapi.SecurityDetectionsAPIRuleInterval(d.Interval.ValueString())
		queryRule.Interval = &interval
	}

	// Set index patterns
	if utils.IsKnown(d.Index) {
		indexList := utils.ListTypeAs[string](ctx, d.Index, path.Root("index"), &diags)
		if !diags.HasError() && len(indexList) > 0 {
			queryRule.Index = &indexList
		}
	}

	// Set author
	if utils.IsKnown(d.Author) {
		authorList := utils.ListTypeAs[string](ctx, d.Author, path.Root("author"), &diags)
		if !diags.HasError() && len(authorList) > 0 {
			queryRule.Author = &authorList
		}
	}

	// Set tags
	if utils.IsKnown(d.Tags) {
		tagsList := utils.ListTypeAs[string](ctx, d.Tags, path.Root("tags"), &diags)
		if !diags.HasError() && len(tagsList) > 0 {
			queryRule.Tags = &tagsList
		}
	}

	// Set false positives
	if utils.IsKnown(d.FalsePositives) {
		fpList := utils.ListTypeAs[string](ctx, d.FalsePositives, path.Root("false_positives"), &diags)
		if !diags.HasError() && len(fpList) > 0 {
			queryRule.FalsePositives = &fpList
		}
	}

	// Set references
	if utils.IsKnown(d.References) {
		refList := utils.ListTypeAs[string](ctx, d.References, path.Root("references"), &diags)
		if !diags.HasError() && len(refList) > 0 {
			queryRule.References = &refList
		}
	}

	// Set optional string fields
	if utils.IsKnown(d.License) {
		license := kbapi.SecurityDetectionsAPIRuleLicense(d.License.ValueString())
		queryRule.License = &license
	}

	if utils.IsKnown(d.Note) {
		note := kbapi.SecurityDetectionsAPIInvestigationGuide(d.Note.ValueString())
		queryRule.Note = &note
	}

	if utils.IsKnown(d.Setup) {
		setup := kbapi.SecurityDetectionsAPISetupGuide(d.Setup.ValueString())
		queryRule.Setup = &setup
	}

	// Set max signals
	if utils.IsKnown(d.MaxSignals) {
		maxSignals := kbapi.SecurityDetectionsAPIMaxSignals(d.MaxSignals.ValueInt64())
		queryRule.MaxSignals = &maxSignals
	}

	// Set version
	if utils.IsKnown(d.Version) {
		version := kbapi.SecurityDetectionsAPIRuleVersion(d.Version.ValueInt64())
		queryRule.Version = &version
	}

	// Convert to union type
	err := createProps.FromSecurityDetectionsAPIQueryRuleCreateProps(queryRule)
	if err != nil {
		diags.AddError(
			"Error building create properties",
			"Could not convert rule properties: "+err.Error(),
		)
	}

	return createProps, diags
}

func (d SecurityDetectionRuleData) toUpdateProps(ctx context.Context) (kbapi.SecurityDetectionsAPIRuleUpdateProps, diag.Diagnostics) {
	var diags diag.Diagnostics
	var updateProps kbapi.SecurityDetectionsAPIRuleUpdateProps

	queryRuleQuery := kbapi.SecurityDetectionsAPIRuleQuery(d.Query.ValueString())

	// Parse ID to get space_id and rule_id
	compId, resourceIdDiags := clients.CompositeIdFromStrFw(d.Id.ValueString())
	diags.Append(resourceIdDiags...)

	uid, err := uuid.Parse(compId.ResourceId)
	if err != nil {
		diags.AddError("ID was not a valid UUID", err.Error())
		return updateProps, diags
	}
	var id = kbapi.SecurityDetectionsAPIRuleObjectId(uid)

	queryRule := kbapi.SecurityDetectionsAPIQueryRuleUpdateProps{
		Id:          &id,
		Name:        kbapi.SecurityDetectionsAPIRuleName(d.Name.ValueString()),
		Description: kbapi.SecurityDetectionsAPIRuleDescription(d.Description.ValueString()),
		Type:        kbapi.SecurityDetectionsAPIQueryRuleUpdatePropsType("query"),
		Query:       &queryRuleQuery,
		RiskScore:   kbapi.SecurityDetectionsAPIRiskScore(d.RiskScore.ValueInt64()),
		Severity:    kbapi.SecurityDetectionsAPISeverity(d.Severity.ValueString()),
	}

	// For updates, we need to include the rule_id if it's set
	if utils.IsKnown(d.RuleId) {
		ruleId := kbapi.SecurityDetectionsAPIRuleSignatureId(d.RuleId.ValueString())
		queryRule.RuleId = &ruleId
		queryRule.Id = nil // if rule_id is set, we cant send id
	}

	// Set enabled status
	if utils.IsKnown(d.Enabled) {
		enabled := kbapi.SecurityDetectionsAPIIsRuleEnabled(d.Enabled.ValueBool())
		queryRule.Enabled = &enabled
	}

	// Set query language
	if utils.IsKnown(d.Language) {
		var language kbapi.SecurityDetectionsAPIKqlQueryLanguage
		switch d.Language.ValueString() {
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
	if utils.IsKnown(d.From) {
		from := kbapi.SecurityDetectionsAPIRuleIntervalFrom(d.From.ValueString())
		queryRule.From = &from
	}

	if utils.IsKnown(d.To) {
		to := kbapi.SecurityDetectionsAPIRuleIntervalTo(d.To.ValueString())
		queryRule.To = &to
	}

	// Set interval
	if utils.IsKnown(d.Interval) {
		interval := kbapi.SecurityDetectionsAPIRuleInterval(d.Interval.ValueString())
		queryRule.Interval = &interval
	}

	// Set index patterns
	if utils.IsKnown(d.Index) {
		indexList := utils.ListTypeAs[string](ctx, d.Index, path.Root("index"), &diags)
		if !diags.HasError() {
			queryRule.Index = &indexList
		}
	}

	// Set author
	if utils.IsKnown(d.Author) {
		authorList := utils.ListTypeAs[string](ctx, d.Author, path.Root("author"), &diags)
		if !diags.HasError() {
			queryRule.Author = &authorList
		}
	}

	// Set tags
	if utils.IsKnown(d.Tags) {
		tagsList := utils.ListTypeAs[string](ctx, d.Tags, path.Root("tags"), &diags)
		if !diags.HasError() {
			queryRule.Tags = &tagsList
		}
	}

	// Set false positives
	if utils.IsKnown(d.FalsePositives) {
		fpList := utils.ListTypeAs[string](ctx, d.FalsePositives, path.Root("false_positives"), &diags)
		if !diags.HasError() {
			queryRule.FalsePositives = &fpList
		}
	}

	// Set references
	if utils.IsKnown(d.References) {
		refList := utils.ListTypeAs[string](ctx, d.References, path.Root("references"), &diags)
		if !diags.HasError() {
			queryRule.References = &refList
		}
	}

	// Set optional string fields
	if utils.IsKnown(d.License) {
		license := kbapi.SecurityDetectionsAPIRuleLicense(d.License.ValueString())
		queryRule.License = &license
	}

	if utils.IsKnown(d.Note) {
		note := kbapi.SecurityDetectionsAPIInvestigationGuide(d.Note.ValueString())
		queryRule.Note = &note
	}

	if utils.IsKnown(d.Setup) {
		setup := kbapi.SecurityDetectionsAPISetupGuide(d.Setup.ValueString())
		queryRule.Setup = &setup
	}

	// Set max signals
	if utils.IsKnown(d.MaxSignals) {
		maxSignals := kbapi.SecurityDetectionsAPIMaxSignals(d.MaxSignals.ValueInt64())
		queryRule.MaxSignals = &maxSignals
	}

	// Set version
	if utils.IsKnown(d.Version) {
		version := kbapi.SecurityDetectionsAPIRuleVersion(d.Version.ValueInt64())
		queryRule.Version = &version
	}

	// Convert to union type
	err = updateProps.FromSecurityDetectionsAPIQueryRuleUpdateProps(queryRule)
	if err != nil {
		diags.AddError(
			"Error building update properties",
			"Could not convert rule properties: "+err.Error(),
		)
	}

	return updateProps, diags
}

func (d *SecurityDetectionRuleData) updateFromRule(ctx context.Context, rule *kbapi.SecurityDetectionsAPIQueryRule) diag.Diagnostics {
	var diags diag.Diagnostics

	compId := clients.CompositeId{
		ClusterId:  d.SpaceId.ValueString(),
		ResourceId: rule.Id.String(),
	}
	d.Id = types.StringValue(compId.String())

	d.RuleId = types.StringValue(string(rule.RuleId))
	d.Name = types.StringValue(string(rule.Name))
	d.Type = types.StringValue(string(rule.Type))
	d.Query = types.StringValue(rule.Query)
	d.Language = types.StringValue(string(rule.Language))
	d.Enabled = types.BoolValue(bool(rule.Enabled))
	d.From = types.StringValue(string(rule.From))
	d.To = types.StringValue(string(rule.To))
	d.Interval = types.StringValue(string(rule.Interval))
	d.Description = types.StringValue(string(rule.Description))
	d.RiskScore = types.Int64Value(int64(rule.RiskScore))
	d.Severity = types.StringValue(string(rule.Severity))
	d.MaxSignals = types.Int64Value(int64(rule.MaxSignals))
	d.Version = types.Int64Value(int64(rule.Version))

	// Update read-only fields
	d.CreatedAt = types.StringValue(rule.CreatedAt.Format("2006-01-02T15:04:05.000Z"))
	d.CreatedBy = types.StringValue(rule.CreatedBy)
	d.UpdatedAt = types.StringValue(rule.UpdatedAt.Format("2006-01-02T15:04:05.000Z"))
	d.UpdatedBy = types.StringValue(rule.UpdatedBy)
	d.Revision = types.Int64Value(int64(rule.Revision))

	// Update index patterns
	if rule.Index != nil && len(*rule.Index) > 0 {
		d.Index = utils.ListValueFrom(ctx, *rule.Index, types.StringType, path.Root("index"), &diags)
	} else {
		d.Index = types.ListValueMust(types.StringType, []attr.Value{})
	}

	// Update author
	if len(rule.Author) > 0 {
		d.Author = utils.ListValueFrom(ctx, rule.Author, types.StringType, path.Root("author"), &diags)
	} else {
		d.Author = types.ListValueMust(types.StringType, []attr.Value{})
	}

	// Update tags
	if len(rule.Tags) > 0 {
		d.Tags = utils.ListValueFrom(ctx, rule.Tags, types.StringType, path.Root("tags"), &diags)
	} else {
		d.Tags = types.ListValueMust(types.StringType, []attr.Value{})
	}

	// Update false positives
	if len(rule.FalsePositives) > 0 {
		d.FalsePositives = utils.ListValueFrom(ctx, rule.FalsePositives, types.StringType, path.Root("false_positives"), &diags)
	} else {
		d.FalsePositives = types.ListValueMust(types.StringType, []attr.Value{})
	}

	// Update references
	if len(rule.References) > 0 {
		d.References = utils.ListValueFrom(ctx, rule.References, types.StringType, path.Root("references"), &diags)
	} else {
		d.References = types.ListValueMust(types.StringType, []attr.Value{})
	}

	// Update optional string fields
	if rule.License != nil {
		d.License = types.StringValue(string(*rule.License))
	} else {
		d.License = types.StringNull()
	}

	if rule.Note != nil {
		d.Note = types.StringValue(string(*rule.Note))
	} else {
		d.Note = types.StringNull()
	}

	// Handle setup field - if empty, set to null to maintain consistency with optional schema
	if string(rule.Setup) != "" {
		d.Setup = types.StringValue(string(rule.Setup))
	} else {
		d.Setup = types.StringNull()
	}

	return diags
}
