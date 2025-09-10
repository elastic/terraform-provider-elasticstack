package prebuilt_rules

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana_oapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type prebuiltRuleModel struct {
	ID                    types.String `tfsdk:"id"`
	SpaceID               types.String `tfsdk:"space_id"`
	Tags                  types.List   `tfsdk:"tags"` // []string
	RulesInstalled        types.Int64  `tfsdk:"rules_installed"`
	RulesNotInstalled     types.Int64  `tfsdk:"rules_not_installed"`
	RulesNotUpdated       types.Int64  `tfsdk:"rules_not_updated"`
	TimelinesInstalled    types.Int64  `tfsdk:"timelines_installed"`
	TimelinesNotInstalled types.Int64  `tfsdk:"timelines_not_installed"`
	TimelinesNotUpdated   types.Int64  `tfsdk:"timelines_not_updated"`
}

type prebuiltRulesStatus struct {
	RulesInstalled        int `json:"rules_installed"`
	RulesNotInstalled     int `json:"rules_not_installed"`
	RulesNotUpdated       int `json:"rules_not_updated"`
	TimelinesInstalled    int `json:"timelines_installed"`
	TimelinesNotInstalled int `json:"timelines_not_installed"`
	TimelinesNotUpdated   int `json:"timelines_not_updated"`
}

func (model *prebuiltRuleModel) populateFromStatus(ctx context.Context, status *prebuiltRulesStatus) diag.Diagnostics {
	if status == nil {
		return nil
	}

	model.RulesInstalled = types.Int64Value(int64(status.RulesInstalled))
	model.RulesNotInstalled = types.Int64Value(int64(status.RulesNotInstalled))
	model.RulesNotUpdated = types.Int64Value(int64(status.RulesNotUpdated))
	model.TimelinesInstalled = types.Int64Value(int64(status.TimelinesInstalled))
	model.TimelinesNotInstalled = types.Int64Value(int64(status.TimelinesNotInstalled))
	model.TimelinesNotUpdated = types.Int64Value(int64(status.TimelinesNotUpdated))

	return nil
}

func (model *prebuiltRuleModel) getTags(ctx context.Context) ([]string, diag.Diagnostics) {
	if model.Tags.IsNull() || model.Tags.IsUnknown() {
		return nil, nil
	}

	var tags []string
	diags := model.Tags.ElementsAs(ctx, &tags, false)
	return tags, diags
}

func getPrebuiltRulesStatus(ctx context.Context, client *kibana_oapi.Client, spaceID string) (*prebuiltRulesStatus, diag.Diagnostics) {
	var resp *kbapi.ReadPrebuiltRulesAndTimelinesStatusResponse
	var err error

	if spaceID != "default" {
		resp, err = client.API.ReadPrebuiltRulesAndTimelinesStatusWithResponse(ctx, func(ctx context.Context, req *http.Request) error {
			req.Header.Set("kbn-space-id", spaceID)
			return nil
		})
	} else {
		resp, err = client.API.ReadPrebuiltRulesAndTimelinesStatusWithResponse(ctx)
	}

	if err != nil {
		return nil, utils.FrameworkDiagFromError(err)
	}

	if resp.StatusCode() != 200 {
		return nil, utils.FrameworkDiagFromError(fmt.Errorf("failed to get prebuilt rules status: %s", resp.Status()))
	}

	var status prebuiltRulesStatus
	if err := json.Unmarshal(resp.Body, &status); err != nil {
		return nil, utils.FrameworkDiagFromError(err)
	}

	return &status, nil
}

func installPrebuiltRules(ctx context.Context, client *kibana_oapi.Client, spaceID string) diag.Diagnostics {
	var resp *kbapi.InstallPrebuiltRulesAndTimelinesResponse
	var err error

	if spaceID != "default" {
		resp, err = client.API.InstallPrebuiltRulesAndTimelinesWithResponse(ctx, func(ctx context.Context, req *http.Request) error {
			req.Header.Set("kbn-space-id", spaceID)
			return nil
		})
	} else {
		resp, err = client.API.InstallPrebuiltRulesAndTimelinesWithResponse(ctx)
	}

	if err != nil {
		return utils.FrameworkDiagFromError(err)
	}

	if resp.StatusCode() != 200 {
		return utils.FrameworkDiagFromError(fmt.Errorf("failed to install prebuilt rules: %s", resp.Status()))
	}

	return nil
}

func needsRuleUpdate(ctx context.Context, client *kibana_oapi.Client, spaceID string) bool {
	status, diags := getPrebuiltRulesStatus(ctx, client, spaceID)
	if diags.HasError() {
		return true
	}
	return status.RulesNotInstalled >= 1 || status.RulesNotUpdated >= 1
}

func manageRulesByTags(ctx context.Context, client *kibana_oapi.Client, spaceID string, tags []string) diag.Diagnostics {
	// Reject "all" as it's not supported by Kibana for large rule sets
	if len(tags) == 1 && tags[0] == "all" {
		return utils.FrameworkDiagFromError(fmt.Errorf("enabling all rules is not supported due to Kibana API limitations. Please specify specific tags to enable a subset of rules"))
	}

	if len(tags) == 0 {
		// If no tags specified, this resource manages no rules - this is valid
		return nil
	}

	// Enable rules matching the specified tags
	diags := performBulkActionByTags(ctx, client, spaceID, "enable", tags)
	if diags.HasError() {
		return diags
	}

	return nil
}

func manageRulesTagTransition(ctx context.Context, client *kibana_oapi.Client, spaceID string, oldTags, newTags []string) diag.Diagnostics {
	// Handle tag transitions for declarative behavior

	// Find tags that were removed (need to disable their rules)
	var removedTags []string
	for _, oldTag := range oldTags {
		found := false
		for _, newTag := range newTags {
			if oldTag == newTag {
				found = true
				break
			}
		}
		if !found {
			removedTags = append(removedTags, oldTag)
		}
	}

	// Find tags that were added (need to enable their rules)
	var addedTags []string
	for _, newTag := range newTags {
		found := false
		for _, oldTag := range oldTags {
			if newTag == oldTag {
				found = true
				break
			}
		}
		if !found {
			addedTags = append(addedTags, newTag)
		}
	}

	// Disable rules for removed tags
	if len(removedTags) > 0 {
		diags := performBulkActionByTags(ctx, client, spaceID, "disable", removedTags)
		if diags.HasError() {
			return diags
		}
	}

	// Enable rules for added tags
	if len(addedTags) > 0 {
		diags := performBulkActionByTags(ctx, client, spaceID, "enable", addedTags)
		if diags.HasError() {
			return diags
		}
	}

	return nil
}

func performBulkActionByTags(ctx context.Context, client *kibana_oapi.Client, spaceID, action string, tags []string) diag.Diagnostics {
	if len(tags) == 0 {
		return nil
	}

	// Reject "all" as it's not supported by Kibana for large rule sets
	if len(tags) == 1 && tags[0] == "all" {
		return utils.FrameworkDiagFromError(fmt.Errorf("enabling all rules is not supported due to Kibana API limitations. Please specify specific tags to enable a subset of rules"))
	}

	// Build KQL query for specific tags
	tagQueries := make([]string, len(tags))
	for i, tag := range tags {
		tagQueries[i] = fmt.Sprintf("alert.attributes.tags:\"%s\"", tag)
	}
	query := strings.Join(tagQueries, " OR ")

	bulkActionBody := map[string]interface{}{
		"action": action,
		"query":  query,
	}

	bodyJSON, err := json.Marshal(bulkActionBody)
	if err != nil {
		return utils.FrameworkDiagFromError(err)
	}

	var resp *kbapi.PerformRulesBulkActionResponse

	if spaceID != "default" {
		resp, err = client.API.PerformRulesBulkActionWithBodyWithResponse(ctx, &kbapi.PerformRulesBulkActionParams{}, "application/json", bytes.NewReader(bodyJSON), func(ctx context.Context, req *http.Request) error {
			req.Header.Set("kbn-space-id", spaceID)
			return nil
		})
	} else {
		resp, err = client.API.PerformRulesBulkActionWithBodyWithResponse(ctx, &kbapi.PerformRulesBulkActionParams{}, "application/json", bytes.NewReader(bodyJSON))
	}

	if err != nil {
		return utils.FrameworkDiagFromError(err)
	}

	if resp.StatusCode() != 200 {
		bodyStr := string(resp.Body)
		return utils.FrameworkDiagFromError(fmt.Errorf("failed to perform bulk %s action on rules: %s. Response: %s", action, resp.Status(), bodyStr))
	}

	return nil
}
