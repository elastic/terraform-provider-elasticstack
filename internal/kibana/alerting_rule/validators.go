package alerting_rule

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// jsonStringValidator validates that a string is valid JSON.
type jsonStringValidator struct{}

func (v jsonStringValidator) Description(_ context.Context) string {
	return "string must be valid JSON"
}

func (v jsonStringValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v jsonStringValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := req.ConfigValue.ValueString()
	var js json.RawMessage
	if err := json.Unmarshal([]byte(value), &js); err != nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid JSON",
			"The value must be a valid JSON string: "+err.Error(),
		)
	}
}

// ValidJSON returns a validator that checks if a string is valid JSON.
func ValidJSON() validator.String {
	return jsonStringValidator{}
}

// ruleTypeParamsValidator is a ConfigValidator that validates params based on rule_type_id.
type ruleTypeParamsValidator struct{}

func (v ruleTypeParamsValidator) Description(_ context.Context) string {
	return "validates that params contain required fields for the specified rule_type_id"
}

func (v ruleTypeParamsValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v ruleTypeParamsValidator) ValidateResource(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	// Get rule_type_id using types.String to handle unknown/null values
	var ruleTypeID types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("rule_type_id"), &ruleTypeID)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get params using types.String to handle unknown/null values
	var params types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("params"), &params)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Skip validation if values are unknown (computed at apply time)
	if ruleTypeID.IsUnknown() || params.IsUnknown() {
		return
	}

	if ruleTypeID.IsNull() || ruleTypeID.ValueString() == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("rule_type_id"),
			"Missing required attribute",
			"The rule_type_id attribute is required and cannot be empty.",
		)
		return
	}

	// Error if params is null or empty
	if params.IsNull() || params.ValueString() == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("params"),
			"Missing required attribute",
			"The params attribute is required and cannot be empty.",
		)
		return
	}

	// Validate params based on rule_type_id
	if err := validateParamsForRuleType(ruleTypeID.ValueString(), params.ValueString()); err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("params"),
			"Invalid params for rule type",
			fmt.Sprintf("The params are invalid for rule type '%s': %s", ruleTypeID.ValueString(), err.Error()),
		)
	}
}

// ValidateRuleTypeParams returns a ConfigValidator that validates params based on rule_type_id.
func ValidateRuleTypeParams() resource.ConfigValidator {
	return ruleTypeParamsValidator{}
}

// ruleTypeToParamsType maps rule_type_id to the corresponding generated params struct type.
// When the OpenAPI spec changes and types are regenerated, update this map.
var ruleTypeToParamsType = map[string]func() interface{}{
	".index-threshold":                      func() interface{} { return &kbapi.ParamsIndexThresholdRule{} },
	"apm.anomaly":                           func() interface{} { return &kbapi.ParamsPropertyApmAnomaly{} },
	"apm.error_rate":                        func() interface{} { return &kbapi.ParamsPropertyApmErrorCount{} },
	"apm.transaction_duration":              func() interface{} { return &kbapi.ParamsPropertyApmTransactionDuration{} },
	"apm.transaction_error_rate":            func() interface{} { return &kbapi.ParamsPropertyApmTransactionErrorRate{} },
	"metrics.alert.inventory.threshold":     func() interface{} { return &kbapi.ParamsPropertyInfraInventory{} },
	"metrics.alert.threshold":               func() interface{} { return &kbapi.ParamsPropertyInfraMetricThreshold{} },
	"slo.rules.burnRate":                    func() interface{} { return &kbapi.ParamsPropertySloBurnRate{} },
	"xpack.uptime.alerts.tls":               func() interface{} { return &kbapi.ParamsPropertySyntheticsUptimeTls{} },
	"xpack.synthetics.alerts.monitorStatus": func() interface{} { return &kbapi.ParamsPropertySyntheticsMonitorStatus{} },
}

// esQuerySearchTypeToParamsType maps .es-query searchType values to their corresponding struct types.
var esQuerySearchTypeToParamsType = map[string]func() interface{}{
	"esQuery":      func() interface{} { return &kbapi.ParamsEsQueryDslRule{} },
	"":             func() interface{} { return &kbapi.ParamsEsQueryDslRule{} }, // default
	"esqlQuery":    func() interface{} { return &kbapi.ParamsEsQueryEsqlRule{} },
	"searchSource": func() interface{} { return &kbapi.ParamsEsQueryKqlRule{} },
}

// validateParamsForRuleType validates the params JSON against the expected schema for the given rule type.
func validateParamsForRuleType(ruleTypeID, paramsJSON string) error {
	// Special handling for .es-query which has sub-types based on searchType
	if ruleTypeID == ".es-query" {
		return validateESQueryParams(paramsJSON)
	}

	// Special handling for log threshold which uses a union type
	if ruleTypeID == "logs.alert.document.count" {
		return validateLogThresholdParams(paramsJSON)
	}

	// Look up the params type for this rule
	paramsTypeFactory, ok := ruleTypeToParamsType[ruleTypeID]
	if !ok {
		// Unknown rule type - just validate it's valid JSON (already done by ValidJSON)
		return nil
	}

	// Create a new instance of the params struct and validate
	paramsStruct := paramsTypeFactory()
	return validateParamsAgainstType(paramsJSON, paramsStruct)
}

// validateESQueryParams validates params for .es-query rules.
// The .es-query rule supports three search types: esQuery (DSL), esqlQuery (ES|QL), and searchSource (KQL).
func validateESQueryParams(paramsJSON string) error {
	// First, determine the search type
	var searchTypeCheck struct {
		SearchType string `json:"searchType"`
	}
	if err := json.Unmarshal([]byte(paramsJSON), &searchTypeCheck); err != nil {
		return fmt.Errorf("failed to parse params: %w", err)
	}

	// Get the appropriate params type factory
	paramsTypeFactory, ok := esQuerySearchTypeToParamsType[searchTypeCheck.SearchType]
	if !ok {
		return fmt.Errorf("unknown searchType: %s (must be 'esQuery', 'esqlQuery', or 'searchSource')", searchTypeCheck.SearchType)
	}

	// Validate against the appropriate struct type
	paramsStruct := paramsTypeFactory()
	return validateParamsAgainstType(paramsJSON, paramsStruct)
}

// validateLogThresholdParams validates params for log threshold rules.
// This uses a union type that requires special handling.
func validateLogThresholdParams(paramsJSON string) error {
	var params kbapi.ParamsPropertyLogThreshold
	if err := params.UnmarshalJSON([]byte(paramsJSON)); err != nil {
		return fmt.Errorf("failed to parse params: %w", err)
	}
	return nil
}

// validateParamsAgainstType unmarshals JSON into the given struct type and validates required fields.
// Required fields are identified as non-pointer struct fields that are not present in the JSON.
// Unknown fields that are not part of the struct are also detected and reported.
func validateParamsAgainstType(paramsJSON string, paramsStruct interface{}) error {
	// First, parse the JSON to get the keys that are actually present
	var rawMap map[string]json.RawMessage
	if err := json.Unmarshal([]byte(paramsJSON), &rawMap); err != nil {
		return fmt.Errorf("failed to parse params: %w", err)
	}

	// Then unmarshal into the struct to validate types
	if err := json.Unmarshal([]byte(paramsJSON), paramsStruct); err != nil {
		return fmt.Errorf("failed to parse params: %w", err)
	}

	// Get valid field names from the struct
	validNames := getValidJSONFieldNames(paramsStruct)

	// Check for unknown fields
	unknown := findUnknownFields(rawMap, validNames)
	if len(unknown) > 0 {
		return fmt.Errorf("unknown fields: %s", strings.Join(unknown, ", "))
	}

	// Use reflection to find missing required fields
	missing := findMissingRequiredFields(paramsStruct, rawMap)
	if len(missing) > 0 {
		return fmt.Errorf("missing required fields: %s", strings.Join(missing, ", "))
	}

	return nil
}

// findMissingRequiredFields uses reflection to find non-pointer fields that are not present in the JSON.
// In the generated kbapi types, required fields are non-pointer types, optional fields are pointers.
func findMissingRequiredFields(v interface{}, presentKeys map[string]json.RawMessage) []string {
	var missing []string

	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return missing
	}

	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		// Skip unexported fields
		if !field.CanInterface() {
			continue
		}

		// Get the JSON field name from the tag
		jsonTag := fieldType.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}
		jsonName := strings.Split(jsonTag, ",")[0]

		// Check if the tag has omitempty - if so, it's optional
		if strings.Contains(jsonTag, "omitempty") {
			continue
		}

		// Check if this is a required field (non-pointer type without omitempty)
		if fieldType.Type.Kind() != reflect.Ptr {
			// Check if the field was present in the JSON
			if _, present := presentKeys[jsonName]; !present {
				missing = append(missing, jsonName)
			}
		}
	}

	return missing
}

// getValidJSONFieldNames extracts all valid JSON field names from a struct using reflection.
func getValidJSONFieldNames(v interface{}) map[string]bool {
	validNames := make(map[string]bool)

	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return validNames
	}

	typ := val.Type()

	for i := 0; i < typ.NumField(); i++ {
		fieldType := typ.Field(i)

		// Get the JSON field name from the tag
		jsonTag := fieldType.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}
		jsonName := strings.Split(jsonTag, ",")[0]
		validNames[jsonName] = true
	}

	return validNames
}

// findUnknownFields finds keys in the JSON that don't exist in the struct's valid field names.
func findUnknownFields(presentKeys map[string]json.RawMessage, validNames map[string]bool) []string {
	var unknown []string
	for key := range presentKeys {
		if !validNames[key] {
			unknown = append(unknown, key)
		}
	}
	return unknown
}
