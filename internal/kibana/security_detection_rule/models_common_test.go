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

package securitydetectionrule

import (
	"reflect"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/stretchr/testify/require"
)

// assertCommonRulePropsWiring checks, for every field of CommonRuleProps, that it was wired
// (non-nil) unless its name is listed in wantAbsent, in which case it must be left nil. This
// guards against a future rule type silently dropping (or a typo silently un-wiring) a common
// field when common_props.go or a rule's model file changes.
func assertCommonRulePropsWiring(t *testing.T, common *CommonRuleProps, wantAbsent ...string) {
	t.Helper()

	absent := make(map[string]struct{}, len(wantAbsent))
	for _, name := range wantAbsent {
		absent[name] = struct{}{}
	}

	v := reflect.ValueOf(*common)
	typ := v.Type()
	for i := range typ.NumField() {
		name := typ.Field(i).Name
		isNil := v.Field(i).IsNil()
		if _, expectAbsent := absent[name]; expectAbsent {
			require.Truef(t, isNil, "field %q: expected to be left nil (not present on source), but it was wired", name)
			continue
		}
		require.Falsef(t, isNil, "field %q: expected to be wired to the source struct, but it was left nil", name)
	}
}

func TestBuildCommonRuleProps(t *testing.T) {
	t.Run("EqlRuleCreateProps", func(t *testing.T) {
		var props kbapi.SecurityDetectionsAPIEqlRuleCreateProps
		assertCommonRulePropsWiring(t, buildCommonRuleProps(&props))
	})
	t.Run("EqlRuleUpdateProps", func(t *testing.T) {
		var props kbapi.SecurityDetectionsAPIEqlRuleUpdateProps
		assertCommonRulePropsWiring(t, buildCommonRuleProps(&props))
	})

	t.Run("EsqlRuleCreateProps", func(t *testing.T) {
		var props kbapi.SecurityDetectionsAPIEsqlRuleCreateProps
		assertCommonRulePropsWiring(t, buildCommonRuleProps(&props), "Index", "DataViewID", "Filters")
	})
	t.Run("EsqlRuleUpdateProps", func(t *testing.T) {
		var props kbapi.SecurityDetectionsAPIEsqlRuleUpdateProps
		assertCommonRulePropsWiring(t, buildCommonRuleProps(&props), "Index", "DataViewID", "Filters")
	})

	t.Run("MachineLearningRuleCreateProps", func(t *testing.T) {
		var props kbapi.SecurityDetectionsAPIMachineLearningRuleCreateProps
		assertCommonRulePropsWiring(t, buildCommonRuleProps(&props), "Index", "DataViewID", "Filters")
	})
	t.Run("MachineLearningRuleUpdateProps", func(t *testing.T) {
		var props kbapi.SecurityDetectionsAPIMachineLearningRuleUpdateProps
		assertCommonRulePropsWiring(t, buildCommonRuleProps(&props), "Index", "DataViewID", "Filters")
	})

	t.Run("NewTermsRuleCreateProps", func(t *testing.T) {
		var props kbapi.SecurityDetectionsAPINewTermsRuleCreateProps
		assertCommonRulePropsWiring(t, buildCommonRuleProps(&props))
	})
	t.Run("NewTermsRuleUpdateProps", func(t *testing.T) {
		var props kbapi.SecurityDetectionsAPINewTermsRuleUpdateProps
		assertCommonRulePropsWiring(t, buildCommonRuleProps(&props))
	})

	t.Run("QueryRuleCreateProps", func(t *testing.T) {
		var props kbapi.SecurityDetectionsAPIQueryRuleCreateProps
		assertCommonRulePropsWiring(t, buildCommonRuleProps(&props))
	})
	t.Run("QueryRuleUpdateProps", func(t *testing.T) {
		var props kbapi.SecurityDetectionsAPIQueryRuleUpdateProps
		assertCommonRulePropsWiring(t, buildCommonRuleProps(&props))
	})

	t.Run("SavedQueryRuleCreateProps", func(t *testing.T) {
		var props kbapi.SecurityDetectionsAPISavedQueryRuleCreateProps
		assertCommonRulePropsWiring(t, buildCommonRuleProps(&props))
	})
	t.Run("SavedQueryRuleUpdateProps", func(t *testing.T) {
		var props kbapi.SecurityDetectionsAPISavedQueryRuleUpdateProps
		assertCommonRulePropsWiring(t, buildCommonRuleProps(&props))
	})

	t.Run("ThreatMatchRuleCreateProps", func(t *testing.T) {
		var props kbapi.SecurityDetectionsAPIThreatMatchRuleCreateProps
		assertCommonRulePropsWiring(t, buildCommonRuleProps(&props))
	})
	t.Run("ThreatMatchRuleUpdateProps", func(t *testing.T) {
		var props kbapi.SecurityDetectionsAPIThreatMatchRuleUpdateProps
		assertCommonRulePropsWiring(t, buildCommonRuleProps(&props))
	})

	t.Run("ThresholdRuleCreateProps", func(t *testing.T) {
		var props kbapi.SecurityDetectionsAPIThresholdRuleCreateProps
		assertCommonRulePropsWiring(t, buildCommonRuleProps(&props, "AlertSuppression"), "AlertSuppression")
	})
	t.Run("ThresholdRuleUpdateProps", func(t *testing.T) {
		var props kbapi.SecurityDetectionsAPIThresholdRuleUpdateProps
		assertCommonRulePropsWiring(t, buildCommonRuleProps(&props, "AlertSuppression"), "AlertSuppression")
	})
}

// TestBuildCommonRuleProps_NameOverrides verifies that fields whose CommonRuleProps name differs
// in casing from the generated struct's field name (RuleID/DataViewID/TimelineID vs
// RuleId/DataViewId/TimelineId) are wired to the correctly-named source field, not silently
// dropped. It does so by mutating through the returned pointer and observing the change on the
// original struct.
func TestBuildCommonRuleProps_NameOverrides(t *testing.T) {
	var props kbapi.SecurityDetectionsAPIEqlRuleCreateProps
	common := buildCommonRuleProps(&props)

	ruleID := kbapi.SecurityDetectionsAPIRuleSignatureId("test-rule-id")
	*common.RuleID = &ruleID
	require.Equal(t, ruleID, *props.RuleId)

	dataViewID := kbapi.SecurityDetectionsAPIDataViewId("test-data-view-id")
	*common.DataViewID = &dataViewID
	require.Equal(t, dataViewID, *props.DataViewId)

	timelineID := kbapi.SecurityDetectionsAPITimelineTemplateId("test-timeline-id")
	*common.TimelineID = &timelineID
	require.Equal(t, timelineID, *props.TimelineId)
}

// assertCommonAPIRuleFieldsWiring checks that buildCommonAPIRuleFields does not panic for the
// given rule type and that fields absent from the source struct are left zero-valued, mirroring
// assertCommonRulePropsWiring but for the API-readback direction. Pointer-typed fields are
// checked explicitly since commonAPIRuleFields mixes pointer and value fields.
func assertCommonAPIRuleFieldsWiring(t *testing.T, fields commonAPIRuleFields, wantAbsent ...string) {
	t.Helper()

	absent := make(map[string]struct{}, len(wantAbsent))
	for _, name := range wantAbsent {
		absent[name] = struct{}{}
	}

	v := reflect.ValueOf(fields)
	typ := v.Type()
	for i := range typ.NumField() {
		field := typ.Field(i)
		if field.Type.Kind() != reflect.Pointer {
			continue
		}
		isNil := v.Field(i).IsNil()
		if _, expectAbsent := absent[field.Name]; expectAbsent {
			require.Truef(t, isNil, "field %q: expected to be left nil (not present on source), but it was set", field.Name)
			continue
		}
		require.Falsef(t, isNil, "field %q: expected to be set from the source struct, but it was left nil", field.Name)
	}
}

func TestBuildCommonAPIRuleFields(t *testing.T) {
	t.Run("EqlRule", func(t *testing.T) {
		rule := newZeroValueCommonAPIRuleFieldsFixture[kbapi.SecurityDetectionsAPIEqlRule]()
		assertCommonAPIRuleFieldsWiring(t, buildCommonAPIRuleFields("id", &rule))
	})

	t.Run("EsqlRule", func(t *testing.T) {
		rule := newZeroValueCommonAPIRuleFieldsFixture[kbapi.SecurityDetectionsAPIEsqlRule]()
		assertCommonAPIRuleFieldsWiring(t, buildCommonAPIRuleFields("id", &rule), "Index", "DataViewID")
	})

	t.Run("MachineLearningRule", func(t *testing.T) {
		rule := newZeroValueCommonAPIRuleFieldsFixture[kbapi.SecurityDetectionsAPIMachineLearningRule]()
		assertCommonAPIRuleFieldsWiring(t, buildCommonAPIRuleFields("id", &rule), "Index", "DataViewID")
	})

	t.Run("NewTermsRule", func(t *testing.T) {
		rule := newZeroValueCommonAPIRuleFieldsFixture[kbapi.SecurityDetectionsAPINewTermsRule]()
		assertCommonAPIRuleFieldsWiring(t, buildCommonAPIRuleFields("id", &rule))
	})

	t.Run("QueryRule", func(t *testing.T) {
		rule := newZeroValueCommonAPIRuleFieldsFixture[kbapi.SecurityDetectionsAPIQueryRule]()
		assertCommonAPIRuleFieldsWiring(t, buildCommonAPIRuleFields("id", &rule))
	})

	t.Run("SavedQueryRule", func(t *testing.T) {
		rule := newZeroValueCommonAPIRuleFieldsFixture[kbapi.SecurityDetectionsAPISavedQueryRule]()
		assertCommonAPIRuleFieldsWiring(t, buildCommonAPIRuleFields("id", &rule))
	})

	t.Run("ThreatMatchRule", func(t *testing.T) {
		rule := newZeroValueCommonAPIRuleFieldsFixture[kbapi.SecurityDetectionsAPIThreatMatchRule]()
		assertCommonAPIRuleFieldsWiring(t, buildCommonAPIRuleFields("id", &rule))
	})

	t.Run("ThresholdRule", func(t *testing.T) {
		rule := newZeroValueCommonAPIRuleFieldsFixture[kbapi.SecurityDetectionsAPIThresholdRule]()
		assertCommonAPIRuleFieldsWiring(t, buildCommonAPIRuleFields("id", &rule, "AlertSuppression"), "AlertSuppression")
	})
}

// newZeroValueCommonAPIRuleFieldsFixture builds a zero-valued API rule struct whose pointer
// fields are allocated (non-nil, pointing at zero values) rather than nil, so the wiring
// assertions above can distinguish "absent from the source struct" (no such field) from "present
// but nil" (a field that exists but the API happened to omit) - both of which are exercised
// elsewhere in the per-rule-type tests in models_test.go using real API payloads.
func newZeroValueCommonAPIRuleFieldsFixture[T any]() T {
	var rule T
	v := reflect.ValueOf(&rule).Elem()
	allocPointerFields(v)
	return rule
}

func allocPointerFields(v reflect.Value) {
	t := v.Type()
	for i := range t.NumField() {
		field := v.Field(i)
		if !field.CanSet() {
			continue
		}
		if field.Kind() == reflect.Pointer && field.IsNil() {
			field.Set(reflect.New(field.Type().Elem()))
		}
	}
}
