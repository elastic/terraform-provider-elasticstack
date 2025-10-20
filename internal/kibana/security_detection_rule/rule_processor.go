package security_detection_rule

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

type ruleProcessor interface {
	HandlesRuleType(t string) bool
	HandlesAPIRuleResponse(rule any) bool
	ToCreateProps(ctx context.Context, client clients.MinVersionEnforceable, d SecurityDetectionRuleData) (kbapi.SecurityDetectionsAPIRuleCreateProps, diag.Diagnostics)
	ToUpdateProps(ctx context.Context, client clients.MinVersionEnforceable, d SecurityDetectionRuleData) (kbapi.SecurityDetectionsAPIRuleUpdateProps, diag.Diagnostics)
	UpdateFromResponse(ctx context.Context, rule any, d *SecurityDetectionRuleData) diag.Diagnostics
	ExtractId(response any) (string, diag.Diagnostics)
}

func getRuleProcessors() []ruleProcessor {
	return []ruleProcessor{
		QueryRuleProcessor{},
		EqlRuleProcessor{},
		EsqlRuleProcessor{},
		MachineLearningRuleProcessor{},
		NewTermsRuleProcessor{},
		SavedQueryRuleProcessor{},
		ThreatMatchRuleProcessor{},
		ThresholdRuleProcessor{},
	}
}

func processorForType(t string) (ruleProcessor, bool) {
	for _, proc := range getRuleProcessors() {
		if proc.HandlesRuleType(t) {
			return proc, true
		}
	}

	return nil, false
}

func getProcessorForResponse(resp *kbapi.SecurityDetectionsAPIRuleResponse) (ruleProcessor, interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics
	respValue, err := resp.ValueByDiscriminator()
	if err != nil {
		diags.AddError(
			"Error determining rule processor",
			"Could not determine the processor for the security detection rule from the API response: "+err.Error(),
		)
		return nil, nil, diags
	}

	for _, proc := range getRuleProcessors() {
		if proc.HandlesAPIRuleResponse(respValue) {
			return proc, respValue, diags
		}
	}

	diags.AddError(
		"Error determining rule processor.",
		"No processor found for rule",
	)

	return nil, nil, diags
}

func (d SecurityDetectionRuleData) toCreateProps(ctx context.Context, client clients.MinVersionEnforceable) (kbapi.SecurityDetectionsAPIRuleCreateProps, diag.Diagnostics) {
	var diags diag.Diagnostics
	var createProps kbapi.SecurityDetectionsAPIRuleCreateProps

	processorForType, ok := processorForType(d.Type.ValueString())
	if !ok {
		diags.AddError(
			"Unsupported rule type",
			fmt.Sprintf("Rule type '%s' is not supported", d.Type.ValueString()),
		)
		return createProps, diags
	}
	return processorForType.ToCreateProps(ctx, client, d)
}

func (d SecurityDetectionRuleData) toUpdateProps(ctx context.Context, client clients.MinVersionEnforceable) (kbapi.SecurityDetectionsAPIRuleUpdateProps, diag.Diagnostics) {
	var diags diag.Diagnostics
	var updateProps kbapi.SecurityDetectionsAPIRuleUpdateProps

	processorForType, ok := processorForType(d.Type.ValueString())
	if !ok {
		diags.AddError(
			"Unsupported rule type",
			fmt.Sprintf("Rule type '%s' is not supported", d.Type.ValueString()),
		)
		return updateProps, diags
	}
	return processorForType.ToUpdateProps(ctx, client, d)
}

func (d *SecurityDetectionRuleData) updateFromRule(ctx context.Context, response *kbapi.SecurityDetectionsAPIRuleResponse) diag.Diagnostics {
	var diags diag.Diagnostics

	// Get the processor for this rule type and use it to update the data
	processorForType, respValue, responseDiags := getProcessorForResponse(response)
	if responseDiags.HasError() {
		diags.Append(responseDiags...)
		return diags
	}

	return processorForType.UpdateFromResponse(ctx, respValue, d)
}

// Helper function to extract rule ID from any rule type
func extractId(response *kbapi.SecurityDetectionsAPIRuleResponse) (string, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Get the processor for this rule type and use it to update the data
	processorForType, respValue, responseDiags := getProcessorForResponse(response)
	if responseDiags.HasError() || processorForType == nil || respValue == nil {
		diags.Append(responseDiags...)
		return "", diags
	}

	return processorForType.ExtractId(respValue)
}
