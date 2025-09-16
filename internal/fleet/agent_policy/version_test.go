package agent_policy

import (
	"context"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"testing"
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
	_, diags = model.toAPIUpdateModel(ctx, feat)
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
	_, diags = model.toAPIUpdateModel(ctx, featSupported)
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
	_, diags = modelWithoutTimeout.toAPIUpdateModel(ctx, feat)
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
	_, diags = model.toAPIUpdateModel(ctx, feat)
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
	_, diags = model.toAPIUpdateModel(ctx, featSupported)
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
	_, diags = modelWithoutTimeout.toAPIUpdateModel(ctx, feat)
	if diags.HasError() {
		t.Errorf("Did not expect error when unenrollment_timeout is not set in update: %v", diags)
	}
}
