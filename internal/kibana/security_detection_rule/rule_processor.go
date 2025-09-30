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
	UpdateFromResponse(ctx context.Context, rule *kbapi.SecurityDetectionsAPIQueryRule, d *SecurityDetectionRuleData) diag.Diagnostics
}

func getRuleProcessors() []ruleProcessor {
	return []ruleProcessor{
		QueryRuleProcessor{},
	}
}

func (d SecurityDetectionRuleData) processorForType(t string) (ruleProcessor, bool) {
	for _, proc := range getRuleProcessors() {
		if proc.HandlesRuleType(t) {
			return proc, true
		}
	}

	return nil, false
}

func (d SecurityDetectionRuleData) getProcessorForResponse(resp *kbapi.SecurityDetectionsAPIRuleResponse) (ruleProcessor, diag.Diagnostics) {
	var diags diag.Diagnostics
	respValue, err := resp.ValueByDiscriminator()
	if err != nil {
		diags.AddError(
			"Error determining rule processor",
			"Could not determine the processor for the security detection rule from the API response: "+err.Error(),
		)
		return nil, diags
	}

	for _, proc := range getRuleProcessors() {
		if proc.HandlesAPIRuleResponse(respValue) {
			return proc, diags
		}
	}

	return nil, diags
}

func (d SecurityDetectionRuleData) toCreateProps(ctx context.Context, client clients.MinVersionEnforceable) (kbapi.SecurityDetectionsAPIRuleCreateProps, diag.Diagnostics) {
	var diags diag.Diagnostics
	var createProps kbapi.SecurityDetectionsAPIRuleCreateProps

	processorForType, ok := d.processorForType(d.Type.ValueString())
	if !ok {
		diags.AddError(
			"Unsupported rule type",
			fmt.Sprintf("Rule type '%s' is not supported", d.Type.ValueString()),
		)
		return createProps, diags
	}
	return processorForType.ToCreateProps(ctx, client, d)
}

func (d *SecurityDetectionRuleData) updateFromRule(ctx context.Context, response *kbapi.SecurityDetectionsAPIRuleResponse) diag.Diagnostics {
	var diags diag.Diagnostics

	rule, err := response.ValueByDiscriminator()
	if err != nil {
		diags.AddError(
			"Error determining rule type",
			"Could not determine the type of the security detection rule from the API response: "+err.Error(),
		)
		return diags
	}

	// Type assertion to check if rule is a SecurityDetectionsAPIQueryRule
	if ruleResp, ok := rule.(kbapi.SecurityDetectionsAPIRuleResponse); ok {
		// Get the processor for this rule type and use it to update the data
		processorForType, err := d.getProcessorForResponse(&ruleResp)
		if err != nil {
			diags.AddError(
				"Error determining rule processor",
				"Could not determine the processor for the security detection rule from the API response: "+err.Error(),
			)
			return diags
		}

		return processorForType.UpdateFromResponse(ctx, d, rule)
	} else {
		diags.AddError(
			"Error determining rule type",
			"Could not determine the type of the security detection rule from the API response"
		)
		return diags
	}
}