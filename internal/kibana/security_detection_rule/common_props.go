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
	"fmt"
	"reflect"
)

// commonFieldNameOverrides maps a CommonRuleProps/commonAPIRuleFields field name to the
// differently-cased field name used on the generated kbapi structs (e.g. RuleID -> RuleId).
var commonFieldNameOverrides = map[string]string{
	"RuleID":     "RuleId",
	"DataViewID": "DataViewId",
	"TimelineID": "TimelineId",
}

// buildCommonRuleProps reflects over props (a pointer to a generated rule Create/UpdateProps
// struct) and wires every CommonRuleProps field to the address of the identically-named field
// on props, replicating what every rule type previously did by hand in a ~30-field struct
// literal. Fields that don't exist on props (e.g. Index/DataViewID/Filters for ESQL and ML
// rules) are left nil, matching the behavior of the former per-rule-type literals. Names
// passed via exclude are always left nil even when present on props (e.g. Threshold's
// AlertSuppression, which is a distinct API type handled separately by the caller).
//
// A field present on both structs under a mismatched type is a programmer error (not a
// runtime condition), so it panics rather than silently dropping the field.
func buildCommonRuleProps[T any](props *T, exclude ...string) *CommonRuleProps {
	skip := make(map[string]struct{}, len(exclude))
	for _, name := range exclude {
		skip[name] = struct{}{}
	}

	var common CommonRuleProps
	commonVal := reflect.ValueOf(&common).Elem()
	commonType := commonVal.Type()
	propsVal := reflect.ValueOf(props).Elem()

	for i := range commonType.NumField() {
		field := commonType.Field(i)
		if _, excluded := skip[field.Name]; excluded {
			continue
		}

		srcName := field.Name
		if override, ok := commonFieldNameOverrides[field.Name]; ok {
			srcName = override
		}

		target := propsVal.FieldByName(srcName)
		if !target.IsValid() {
			continue
		}

		addr := target.Addr()
		if addr.Type() != field.Type {
			panic(fmt.Sprintf("securitydetectionrule: common field %q type mismatch: want %s, got %s on %T", field.Name, field.Type, addr.Type(), props))
		}

		commonVal.Field(i).Set(addr)
	}

	return &common
}

// buildCommonAPIRuleFields reflects over rule (a pointer to a generated rule API response
// struct) and copies every commonAPIRuleFields field from the identically-named field on rule,
// converting between the generated type aliases/enums and the plain Go types commonAPIRuleFields
// uses (e.g. the named SecurityDetectionsAPISeverity -> string, or int -> int64), replicating
// what every rule type previously did by hand in a ~30-field struct literal. ResourceID has no
// equivalent field (it's derived from rule.Id.String()) and must be passed explicitly. Fields
// that don't exist on rule (e.g. Index/DataViewID for ESQL and ML rules) are left zero-valued,
// matching the behavior of the former per-rule-type literals. Names passed via exclude are
// always left zero-valued even when present on rule (e.g. Threshold's AlertSuppression, which
// is a distinct API type handled separately by the caller).
//
// A field present on both structs under a non-convertible type is a programmer error (not a
// runtime condition), so it panics rather than silently dropping the field.
func buildCommonAPIRuleFields[T any](resourceID string, rule *T, exclude ...string) commonAPIRuleFields {
	skip := make(map[string]struct{}, len(exclude))
	for _, name := range exclude {
		skip[name] = struct{}{}
	}

	var fields commonAPIRuleFields
	fields.ResourceID = resourceID

	fieldsVal := reflect.ValueOf(&fields).Elem()
	fieldsType := fieldsVal.Type()
	ruleVal := reflect.ValueOf(rule).Elem()

	for i := range fieldsType.NumField() {
		field := fieldsType.Field(i)
		if field.Name == "ResourceID" {
			continue
		}
		if _, excluded := skip[field.Name]; excluded {
			continue
		}

		srcName := field.Name
		if override, ok := commonFieldNameOverrides[field.Name]; ok {
			srcName = override
		}

		source := ruleVal.FieldByName(srcName)
		if !source.IsValid() {
			continue
		}

		if !source.Type().ConvertibleTo(field.Type) {
			panic(fmt.Sprintf("securitydetectionrule: common API field %q type mismatch: want %s, got %s on %T", field.Name, field.Type, source.Type(), rule))
		}

		fieldsVal.Field(i).Set(source.Convert(field.Type))
	}

	return fields
}
