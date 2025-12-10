package agent_policy

import (
	"context"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestMinVersionInactivityTimeout(t *testing.T) {
	// Test that the MinVersionInactivityTimeout constant is set correctly
	expected := "8.7.0"
	actual := MinVersionInactivityTimeout.String()
	if actual != expected {
		t.Errorf("Expected MinVersionInactivityTimeout to be '%s', got '%s'", expected, actual)
	}

	// Test version comparison - should be greater than 8.6.0
	olderVersion := version.Must(version.NewVersion("8.6.0"))
	if MinVersionInactivityTimeout.LessThan(olderVersion) {
		t.Errorf("MinVersionInactivityTimeout (%s) should be greater than %s", MinVersionInactivityTimeout.String(), olderVersion.String())
	}

	// Test version comparison - should be less than 8.8.0
	newerVersion := version.Must(version.NewVersion("8.8.0"))
	if MinVersionInactivityTimeout.GreaterThan(newerVersion) {
		t.Errorf("MinVersionInactivityTimeout (%s) should be less than %s", MinVersionInactivityTimeout.String(), newerVersion.String())
	}
}

func TestMinVersionUnenrollmentTimeout(t *testing.T) {
	// Test that the MinVersionUnenrollmentTimeout constant is set correctly
	expected := "8.15.0"
	actual := MinVersionUnenrollmentTimeout.String()
	if actual != expected {
		t.Errorf("Expected MinVersionUnenrollmentTimeout to be '%s', got '%s'", expected, actual)
	}

	// Test version comparison - should be greater than 8.14.0
	olderVersion := version.Must(version.NewVersion("8.14.0"))
	if MinVersionUnenrollmentTimeout.LessThan(olderVersion) {
		t.Errorf("MinVersionUnenrollmentTimeout (%s) should be greater than %s", MinVersionUnenrollmentTimeout.String(), olderVersion.String())
	}

	// Test version comparison - should be less than 8.16.0
	newerVersion := version.Must(version.NewVersion("8.16.0"))
	if MinVersionUnenrollmentTimeout.GreaterThan(newerVersion) {
		t.Errorf("MinVersionUnenrollmentTimeout (%s) should be less than %s", MinVersionUnenrollmentTimeout.String(), newerVersion.String())
	}
}

func TestInactivityTimeoutVersionValidation(t *testing.T) {
	ctx := context.Background()

	// Test case where inactivity_timeout is not supported (older version)
	model := &agentPolicyModel{
		Name:              types.StringValue("test"),
		Namespace:         types.StringValue("default"),
		InactivityTimeout: customtypes.NewDurationValue("2m"),
	}

	// Create features with inactivity timeout NOT supported
	feat := features{
		SupportsInactivityTimeout: false,
	}

	// Test toAPICreateModel - should return error when inactivity_timeout is used but not supported
	_, diags := model.toAPICreateModel(ctx, feat)
	if !diags.HasError() {
		t.Error("Expected error when using inactivity_timeout on unsupported version, but got none")
	}

	// Check that the error message contains the expected text
	found := false
	for _, diag := range diags {
		if diag.Summary() == "Unsupported Elasticsearch version" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected 'Unsupported Elasticsearch version' error, but didn't find it")
	}

	// Test toAPIUpdateModel - should return error when inactivity_timeout is used but not supported
	_, diags = model.toAPIUpdateModel(ctx, feat, nil)
	if !diags.HasError() {
		t.Error("Expected error when using inactivity_timeout on unsupported version in update, but got none")
	}

	// Test case where inactivity_timeout IS supported (newer version)
	featSupported := features{
		SupportsInactivityTimeout: true,
	}

	// Test toAPICreateModel - should NOT return error when inactivity_timeout is supported
	_, diags = model.toAPICreateModel(ctx, featSupported)
	if diags.HasError() {
		t.Errorf("Did not expect error when using inactivity_timeout on supported version: %v", diags)
	}

	// Test toAPIUpdateModel - should NOT return error when inactivity_timeout is supported
	_, diags = model.toAPIUpdateModel(ctx, featSupported, nil)
	if diags.HasError() {
		t.Errorf("Did not expect error when using inactivity_timeout on supported version in update: %v", diags)
	}

	// Test case where inactivity_timeout is not set (should not cause validation errors)
	modelWithoutTimeout := &agentPolicyModel{
		Name:      types.StringValue("test"),
		Namespace: types.StringValue("default"),
		// InactivityTimeout is not set (null/unknown)
	}

	// Test toAPICreateModel - should NOT return error when inactivity_timeout is not set, even on unsupported version
	_, diags = modelWithoutTimeout.toAPICreateModel(ctx, feat)
	if diags.HasError() {
		t.Errorf("Did not expect error when inactivity_timeout is not set: %v", diags)
	}

	// Test toAPIUpdateModel - should NOT return error when inactivity_timeout is not set, even on unsupported version
	_, diags = modelWithoutTimeout.toAPIUpdateModel(ctx, feat, nil)
	if diags.HasError() {
		t.Errorf("Did not expect error when inactivity_timeout is not set in update: %v", diags)
	}
}

func TestUnenrollmentTimeoutVersionValidation(t *testing.T) {
	ctx := context.Background()

	// Test case where unenrollment_timeout is not supported (older version)
	model := &agentPolicyModel{
		Name:                types.StringValue("test"),
		Namespace:           types.StringValue("default"),
		UnenrollmentTimeout: customtypes.NewDurationValue("5m"),
	}

	// Create features with unenrollment timeout NOT supported
	feat := features{
		SupportsUnenrollmentTimeout: false,
	}

	// Test toAPICreateModel - should return error when unenrollment_timeout is used but not supported
	_, diags := model.toAPICreateModel(ctx, feat)
	if !diags.HasError() {
		t.Error("Expected error when using unenrollment_timeout on unsupported version, but got none")
	}

	// Check that the error message contains the expected text
	found := false
	for _, diag := range diags {
		if diag.Summary() == "Unsupported Elasticsearch version" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected 'Unsupported Elasticsearch version' error, but didn't find it")
	}

	// Test toAPIUpdateModel - should return error when unenrollment_timeout is used but not supported
	_, diags = model.toAPIUpdateModel(ctx, feat, nil)
	if !diags.HasError() {
		t.Error("Expected error when using unenrollment_timeout on unsupported version in update, but got none")
	}

	// Test case where unenrollment_timeout IS supported (newer version)
	featSupported := features{
		SupportsUnenrollmentTimeout: true,
	}

	// Test toAPICreateModel - should NOT return error when unenrollment_timeout is supported
	_, diags = model.toAPICreateModel(ctx, featSupported)
	if diags.HasError() {
		t.Errorf("Did not expect error when using unenrollment_timeout on supported version: %v", diags)
	}

	// Test toAPIUpdateModel - should NOT return error when unenrollment_timeout is supported
	_, diags = model.toAPIUpdateModel(ctx, featSupported, nil)
	if diags.HasError() {
		t.Errorf("Did not expect error when using unenrollment_timeout on supported version in update: %v", diags)
	}

	// Test case where unenrollment_timeout is not set (should not cause validation errors)
	modelWithoutTimeout := &agentPolicyModel{
		Name:      types.StringValue("test"),
		Namespace: types.StringValue("default"),
		// UnenrollmentTimeout is not set (null/unknown)
	}

	// Test toAPICreateModel - should NOT return error when unenrollment_timeout is not set, even on unsupported version
	_, diags = modelWithoutTimeout.toAPICreateModel(ctx, feat)
	if diags.HasError() {
		t.Errorf("Did not expect error when unenrollment_timeout is not set: %v", diags)
	}

	// Test toAPIUpdateModel - should NOT return error when unenrollment_timeout is not set, even on unsupported version
	_, diags = modelWithoutTimeout.toAPIUpdateModel(ctx, feat, nil)
	if diags.HasError() {
		t.Errorf("Did not expect error when unenrollment_timeout is not set in update: %v", diags)
	}
}

func TestMinVersionSpaceIds(t *testing.T) {
	// Test that the MinVersionSpaceIds constant is set correctly
	expected := "9.1.0"
	actual := MinVersionSpaceIds.String()
	if actual != expected {
		t.Errorf("Expected MinVersionSpaceIds to be '%s', got '%s'", expected, actual)
	}

	// Test version comparison - should be greater than 9.0.0
	olderVersion := version.Must(version.NewVersion("9.0.0"))
	if MinVersionSpaceIds.LessThan(olderVersion) {
		t.Errorf("MinVersionSpaceIds (%s) should be greater than %s", MinVersionSpaceIds.String(), olderVersion.String())
	}

	// Test version comparison - should be less than 9.2.0
	newerVersion := version.Must(version.NewVersion("9.2.0"))
	if MinVersionSpaceIds.GreaterThan(newerVersion) {
		t.Errorf("MinVersionSpaceIds (%s) should be less than %s", MinVersionSpaceIds.String(), newerVersion.String())
	}
}

func TestSpaceIdsVersionValidation(t *testing.T) {
	ctx := context.Background()

	// Test case where space_ids is not supported (older version)
	spaceIds, _ := types.SetValueFrom(ctx, types.StringType, []string{"default", "marketing"})
	model := &agentPolicyModel{
		Name:      types.StringValue("test"),
		Namespace: types.StringValue("default"),
		SpaceIds:  spaceIds,
	}

	// Create features with space_ids NOT supported
	feat := features{
		SupportsSpaceIds: false,
	}

	// Test toAPICreateModel - should return error when space_ids is used but not supported
	_, diags := model.toAPICreateModel(ctx, feat)
	if !diags.HasError() {
		t.Error("Expected error when using space_ids on unsupported version, but got none")
	}

	// Check that the error message contains the expected text
	found := false
	for _, diag := range diags {
		if diag.Summary() == "Unsupported Elasticsearch version" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected 'Unsupported Elasticsearch version' error, but didn't find it")
	}

	// Test toAPIUpdateModel - should return error when space_ids is used but not supported
	_, diags = model.toAPIUpdateModel(ctx, feat, nil)
	if !diags.HasError() {
		t.Error("Expected error when using space_ids on unsupported version in update, but got none")
	}

	// Test case where space_ids IS supported (newer version)
	featSupported := features{
		SupportsSpaceIds: true,
	}

	// Test toAPICreateModel - should NOT return error when space_ids is supported
	_, diags = model.toAPICreateModel(ctx, featSupported)
	if diags.HasError() {
		t.Errorf("Did not expect error when using space_ids on supported version: %v", diags)
	}

	// Test toAPIUpdateModel - should NOT return error when space_ids is supported
	_, diags = model.toAPIUpdateModel(ctx, featSupported, nil)
	if diags.HasError() {
		t.Errorf("Did not expect error when using space_ids on supported version in update: %v", diags)
	}

	// Test case where space_ids is not set (should not cause validation errors)
	modelWithoutSpaceIds := &agentPolicyModel{
		Name:      types.StringValue("test"),
		Namespace: types.StringValue("default"),
		// SpaceIds is not set (null/unknown)
	}

	// Test toAPICreateModel - should NOT return error when space_ids is not set, even on unsupported version
	_, diags = modelWithoutSpaceIds.toAPICreateModel(ctx, feat)
	if diags.HasError() {
		t.Errorf("Did not expect error when space_ids is not set: %v", diags)
	}

	// Test toAPIUpdateModel - should NOT return error when space_ids is not set, even on unsupported version
	_, diags = modelWithoutSpaceIds.toAPIUpdateModel(ctx, feat, nil)
	if diags.HasError() {
		t.Errorf("Did not expect error when space_ids is not set in update: %v", diags)
	}
}

func TestMinVersionAgentFeatures(t *testing.T) {
	// Test that the MinVersionAgentFeatures constant is set correctly
	expected := "8.7.0"
	actual := MinVersionAgentFeatures.String()
	if actual != expected {
		t.Errorf("Expected MinVersionAgentFeatures to be '%s', got '%s'", expected, actual)
	}

	// Test version comparison - should be greater than 8.6.0
	olderVersion := version.Must(version.NewVersion("8.6.0"))
	if MinVersionAgentFeatures.LessThan(olderVersion) {
		t.Errorf("MinVersionAgentFeatures (%s) should be greater than %s", MinVersionAgentFeatures.String(), olderVersion.String())
	}

	// Test version comparison - should be less than 8.8.0
	newerVersion := version.Must(version.NewVersion("8.8.0"))
	if MinVersionAgentFeatures.GreaterThan(newerVersion) {
		t.Errorf("MinVersionAgentFeatures (%s) should be less than %s", MinVersionAgentFeatures.String(), newerVersion.String())
	}
}

func TestAgentFeaturesVersionValidation(t *testing.T) {
	ctx := context.Background()

	// Test case where agent_features is not supported (older version) with FQDN
	model := &agentPolicyModel{
		Name:           types.StringValue("test"),
		Namespace:      types.StringValue("default"),
		HostNameFormat: types.StringValue(HostNameFormatFQDN),
	}

	// Create features with agent_features NOT supported
	feat := features{
		SupportsAgentFeatures: false,
	}

	// Test toAPICreateModel - should return error when host_name_format=fqdn on unsupported version
	_, diags := model.toAPICreateModel(ctx, feat)
	if !diags.HasError() {
		t.Error("Expected error when using host_name_format=fqdn on unsupported version, but got none")
	}

	// Check that the error message contains the expected text
	found := false
	for _, diag := range diags {
		if diag.Summary() == "Unsupported Elasticsearch version" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected 'Unsupported Elasticsearch version' error, but didn't find it")
	}

	// Test toAPIUpdateModel - should return error when host_name_format=fqdn on unsupported version
	_, diags = model.toAPIUpdateModel(ctx, feat, nil)
	if !diags.HasError() {
		t.Error("Expected error when using host_name_format=fqdn on unsupported version in update, but got none")
	}

	// Test case where host_name_format=hostname (default) on unsupported version - should NOT error
	modelWithHostname := &agentPolicyModel{
		Name:           types.StringValue("test"),
		Namespace:      types.StringValue("default"),
		HostNameFormat: types.StringValue(HostNameFormatHostname),
	}

	// Test toAPICreateModel - should NOT return error for hostname (default) on unsupported version
	_, diags = modelWithHostname.toAPICreateModel(ctx, feat)
	if diags.HasError() {
		t.Errorf("Did not expect error when using host_name_format=hostname on unsupported version: %v", diags)
	}

	// Test toAPIUpdateModel - should NOT return error for hostname (default) on unsupported version
	_, diags = modelWithHostname.toAPIUpdateModel(ctx, feat, nil)
	if diags.HasError() {
		t.Errorf("Did not expect error when using host_name_format=hostname on unsupported version in update: %v", diags)
	}

	// Test case where agent_features IS supported (newer version)
	featSupported := features{
		SupportsAgentFeatures: true,
	}

	// Test toAPICreateModel - should NOT return error when agent_features is supported
	_, diags = model.toAPICreateModel(ctx, featSupported)
	if diags.HasError() {
		t.Errorf("Did not expect error when using host_name_format on supported version: %v", diags)
	}

	// Test toAPIUpdateModel - should NOT return error when agent_features is supported
	_, diags = model.toAPIUpdateModel(ctx, featSupported, nil)
	if diags.HasError() {
		t.Errorf("Did not expect error when using host_name_format on supported version in update: %v", diags)
	}

	// Test case where host_name_format is not set (should not cause validation errors)
	modelWithoutHostNameFormat := &agentPolicyModel{
		Name:      types.StringValue("test"),
		Namespace: types.StringValue("default"),
		// HostNameFormat is not set (null/unknown)
	}

	// Test toAPICreateModel - should NOT return error when host_name_format is not set, even on unsupported version
	_, diags = modelWithoutHostNameFormat.toAPICreateModel(ctx, feat)
	if diags.HasError() {
		t.Errorf("Did not expect error when host_name_format is not set: %v", diags)
	}

	// Test toAPIUpdateModel - should NOT return error when host_name_format is not set, even on unsupported version
	_, diags = modelWithoutHostNameFormat.toAPIUpdateModel(ctx, feat, nil)
	if diags.HasError() {
		t.Errorf("Did not expect error when host_name_format is not set in update: %v", diags)
	}
}
