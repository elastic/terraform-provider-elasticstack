// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package elasticdefendintegrationpolicy

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func TestResourceSchemaModelsDefendPolicyDefaults(t *testing.T) {
	resourceSchema := resourceSchema()
	policy := mustSingleNestedAttribute(t, resourceSchema.Attributes, "policy")
	windows := mustSingleNestedAttribute(t, policy.Attributes, "windows")
	mac := mustSingleNestedAttribute(t, policy.Attributes, "mac")
	linux := mustSingleNestedAttribute(t, policy.Attributes, "linux")

	assertPopupDefaults(t, mustSingleNestedAttribute(t, mustSingleNestedAttribute(t, windows.Attributes, "popup").Attributes, "malware"))
	assertPopupDefaults(t, mustSingleNestedAttribute(t, mustSingleNestedAttribute(t, mac.Attributes, "popup").Attributes, "malware"))
	assertPopupDefaults(t, mustSingleNestedAttribute(t, mustSingleNestedAttribute(t, linux.Attributes, "popup").Attributes, "malware"))

	windowsPopup := mustSingleNestedAttribute(t, windows.Attributes, "popup")
	assertHasObjectDefault(t, windowsPopup, "policy.windows.popup")

	assertProtectionModeDefaults(t, mustSingleNestedAttribute(t, windows.Attributes, "ransomware"))
	assertProtectionModeDefaults(t, mustSingleNestedAttribute(t, windows.Attributes, "memory_protection"))
	assertProtectionModeDefaults(t, mustSingleNestedAttribute(t, mac.Attributes, "memory_protection"))
	assertProtectionModeDefaults(t, mustSingleNestedAttribute(t, linux.Attributes, "memory_protection"))

	assertBehaviorProtectionDefaults(t, mustSingleNestedAttribute(t, windows.Attributes, "behavior_protection"))
	assertBehaviorProtectionDefaults(t, mustSingleNestedAttribute(t, mac.Attributes, "behavior_protection"))
	assertBehaviorProtectionDefaults(t, mustSingleNestedAttribute(t, linux.Attributes, "behavior_protection"))

	antivirusRegistration := mustSingleNestedAttribute(t, windows.Attributes, "antivirus_registration")
	assertHasObjectDefault(t, antivirusRegistration, "policy.windows.antivirus_registration")
	assertHasStringDefault(t, mustStringAttribute(t, antivirusRegistration.Attributes, "mode"), "policy.windows.antivirus_registration.mode")
	assertHasBoolDefault(t, mustBoolAttribute(t, antivirusRegistration.Attributes, "enabled"), "policy.windows.antivirus_registration.enabled")

	attackSurfaceReduction := mustSingleNestedAttribute(t, windows.Attributes, "attack_surface_reduction")
	assertHasObjectDefault(t, attackSurfaceReduction, "policy.windows.attack_surface_reduction")

	credentialHardening := mustSingleNestedAttribute(t, attackSurfaceReduction.Attributes, "credential_hardening")
	assertHasObjectDefault(t, credentialHardening, "policy.windows.attack_surface_reduction.credential_hardening")
	assertHasBoolDefault(t, mustBoolAttribute(t, credentialHardening.Attributes, "enabled"), "policy.windows.attack_surface_reduction.credential_hardening.enabled")
}

func assertPopupDefaults(t *testing.T, attr schema.SingleNestedAttribute) {
	t.Helper()
	assertHasObjectDefault(t, attr, "popup item")
	assertHasStringDefault(t, mustStringAttribute(t, attr.Attributes, "message"), "popup item.message")
	assertHasBoolDefault(t, mustBoolAttribute(t, attr.Attributes, "enabled"), "popup item.enabled")
}

func assertProtectionModeDefaults(t *testing.T, attr schema.SingleNestedAttribute) {
	t.Helper()
	assertHasObjectDefault(t, attr, "protection mode")
	assertHasStringDefault(t, mustStringAttribute(t, attr.Attributes, "mode"), "protection mode.mode")
	assertHasBoolDefault(t, mustBoolAttribute(t, attr.Attributes, "supported"), "protection mode.supported")
}

func assertBehaviorProtectionDefaults(t *testing.T, attr schema.SingleNestedAttribute) {
	t.Helper()
	assertHasObjectDefault(t, attr, "behavior protection")
	assertHasStringDefault(t, mustStringAttribute(t, attr.Attributes, "mode"), "behavior protection.mode")
	assertHasBoolDefault(t, mustBoolAttribute(t, attr.Attributes, "supported"), "behavior protection.supported")
	assertHasBoolDefault(t, mustBoolAttribute(t, attr.Attributes, "reputation_service"), "behavior protection.reputation_service")
}

func assertHasObjectDefault(t *testing.T, attr schema.SingleNestedAttribute, path string) {
	t.Helper()
	if !attr.Computed {
		t.Fatalf("%s should be computed when a default is modeled", path)
	}
	if attr.Default == nil {
		t.Fatalf("%s should define an object default", path)
	}
}

func assertHasStringDefault(t *testing.T, attr schema.StringAttribute, path string) {
	t.Helper()
	if !attr.Computed {
		t.Fatalf("%s should be computed when a default is modeled", path)
	}
	if attr.Default == nil {
		t.Fatalf("%s should define a string default", path)
	}
}

func assertHasBoolDefault(t *testing.T, attr schema.BoolAttribute, path string) {
	t.Helper()
	if !attr.Computed {
		t.Fatalf("%s should be computed when a default is modeled", path)
	}
	if attr.Default == nil {
		t.Fatalf("%s should define a bool default", path)
	}
}

func mustSingleNestedAttribute(t *testing.T, attrs map[string]schema.Attribute, name string) schema.SingleNestedAttribute {
	t.Helper()
	attr, ok := attrs[name]
	if !ok {
		t.Fatalf("expected attribute %q to exist", name)
	}
	nestedAttr, ok := attr.(schema.SingleNestedAttribute)
	if !ok {
		t.Fatalf("expected %q to be a SingleNestedAttribute, got %T", name, attr)
	}
	return nestedAttr
}

func mustStringAttribute(t *testing.T, attrs map[string]schema.Attribute, name string) schema.StringAttribute {
	t.Helper()
	attr, ok := attrs[name]
	if !ok {
		t.Fatalf("expected attribute %q to exist", name)
	}
	stringAttr, ok := attr.(schema.StringAttribute)
	if !ok {
		t.Fatalf("expected %q to be a StringAttribute, got %T", name, attr)
	}
	return stringAttr
}

func mustBoolAttribute(t *testing.T, attrs map[string]schema.Attribute, name string) schema.BoolAttribute {
	t.Helper()
	attr, ok := attrs[name]
	if !ok {
		t.Fatalf("expected attribute %q to exist", name)
	}
	boolAttr, ok := attr.(schema.BoolAttribute)
	if !ok {
		t.Fatalf("expected %q to be a BoolAttribute, got %T", name, attr)
	}
	return boolAttr
}
