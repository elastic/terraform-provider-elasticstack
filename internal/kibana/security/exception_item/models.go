package exception_item

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type ExceptionItemModel struct {
	ID            types.String `tfsdk:"id"`
	SpaceID       types.String `tfsdk:"space_id"`
	ItemID        types.String `tfsdk:"item_id"`
	ListID        types.String `tfsdk:"list_id"`
	Name          types.String `tfsdk:"name"`
	Description   types.String `tfsdk:"description"`
	Type          types.String `tfsdk:"type"`
	NamespaceType types.String `tfsdk:"namespace_type"`
	OsTypes       types.List   `tfsdk:"os_types"`
	Tags          types.List   `tfsdk:"tags"`
	Meta          types.String `tfsdk:"meta"`
	Entries       types.List   `tfsdk:"entries"`
	Comments      types.List   `tfsdk:"comments"`
	ExpireTime    types.String `tfsdk:"expire_time"`
	CreatedAt     types.String `tfsdk:"created_at"`
	CreatedBy     types.String `tfsdk:"created_by"`
	UpdatedAt     types.String `tfsdk:"updated_at"`
	UpdatedBy     types.String `tfsdk:"updated_by"`
	TieBreakerID  types.String `tfsdk:"tie_breaker_id"`
}

type CommentModel struct {
	ID      types.String `tfsdk:"id"`
	Comment types.String `tfsdk:"comment"`
}

type EntryModel struct {
	Type     types.String `tfsdk:"type"`
	Field    types.String `tfsdk:"field"`
	Operator types.String `tfsdk:"operator"`
	Value    types.String `tfsdk:"value"`
	Values   types.List   `tfsdk:"values"`
	List     types.Object `tfsdk:"list"`
	Entries  types.List   `tfsdk:"entries"`
}

type EntryListModel struct {
	ID   types.String `tfsdk:"id"`
	Type types.String `tfsdk:"type"`
}

type NestedEntryModel struct {
	Type     types.String `tfsdk:"type"`
	Field    types.String `tfsdk:"field"`
	Operator types.String `tfsdk:"operator"`
	Value    types.String `tfsdk:"value"`
	Values   types.List   `tfsdk:"values"`
}

// convertEntriesToAPI converts Terraform entry models to API entry models
func convertEntriesToAPI(ctx context.Context, entries types.List) (kbapi.SecurityExceptionsAPIExceptionListItemEntryArray, diag.Diagnostics) {
	var diags diag.Diagnostics

	if entries.IsNull() || entries.IsUnknown() {
		return nil, diags
	}

	var entryModels []EntryModel
	diags.Append(entries.ElementsAs(ctx, &entryModels, false)...)
	if diags.HasError() {
		return nil, diags
	}

	apiEntries := make(kbapi.SecurityExceptionsAPIExceptionListItemEntryArray, 0, len(entryModels))
	for _, entry := range entryModels {
		apiEntry, d := convertEntryToAPI(ctx, entry)
		diags.Append(d...)
		if d.HasError() {
			continue
		}
		apiEntries = append(apiEntries, apiEntry)
	}

	return apiEntries, diags
}

// convertEntryToAPI converts a single Terraform entry model to an API entry model
func convertEntryToAPI(ctx context.Context, entry EntryModel) (kbapi.SecurityExceptionsAPIExceptionListItemEntry, diag.Diagnostics) {
	var diags diag.Diagnostics
	var result kbapi.SecurityExceptionsAPIExceptionListItemEntry

	entryType := entry.Type.ValueString()
	operator := kbapi.SecurityExceptionsAPIExceptionListItemEntryOperator(entry.Operator.ValueString())
	field := kbapi.SecurityExceptionsAPINonEmptyString(entry.Field.ValueString())

	switch entryType {
	case "match":
		// Validate required field
		if entry.Value.IsNull() || entry.Value.IsUnknown() || entry.Value.ValueString() == "" {
			diags.AddError("Invalid Configuration", "Attribute 'value' is required when type is 'match'")
			return result, diags
		}

		apiEntry := kbapi.SecurityExceptionsAPIExceptionListItemEntryMatch{
			Type:     "match",
			Field:    field,
			Operator: operator,
			Value:    kbapi.SecurityExceptionsAPINonEmptyString(entry.Value.ValueString()),
		}
		if err := result.FromSecurityExceptionsAPIExceptionListItemEntryMatch(apiEntry); err != nil {
			diags.AddError("Failed to create match entry", err.Error())
		}

	case "match_any":
		// Validate required field
		if entry.Values.IsNull() || entry.Values.IsUnknown() {
			diags.AddError("Invalid Configuration", "Attribute 'values' is required when type is 'match_any'")
			return result, diags
		}

		var values []string
		diags.Append(entry.Values.ElementsAs(ctx, &values, false)...)
		if diags.HasError() {
			return result, diags
		}

		if len(values) == 0 {
			diags.AddError("Invalid Configuration", "Attribute 'values' must contain at least one value when type is 'match_any'")
			return result, diags
		}

		apiValues := make([]kbapi.SecurityExceptionsAPINonEmptyString, len(values))
		for i, v := range values {
			apiValues[i] = kbapi.SecurityExceptionsAPINonEmptyString(v)
		}
		apiEntry := kbapi.SecurityExceptionsAPIExceptionListItemEntryMatchAny{
			Type:     "match_any",
			Field:    field,
			Operator: operator,
			Value:    apiValues,
		}
		if err := result.FromSecurityExceptionsAPIExceptionListItemEntryMatchAny(apiEntry); err != nil {
			diags.AddError("Failed to create match_any entry", err.Error())
		}

	case "list":
		// Validate required field
		if entry.List.IsNull() || entry.List.IsUnknown() {
			diags.AddError("Invalid Configuration", "Attribute 'list' is required when type is 'list'")
			return result, diags
		}

		var listModel EntryListModel
		diags.Append(entry.List.As(ctx, &listModel, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return result, diags
		}
		apiEntry := kbapi.SecurityExceptionsAPIExceptionListItemEntryList{
			Type:     "list",
			Field:    field,
			Operator: operator,
		}
		apiEntry.List.Id = kbapi.SecurityExceptionsAPIListId(listModel.ID.ValueString())
		apiEntry.List.Type = kbapi.SecurityExceptionsAPIListType(listModel.Type.ValueString())
		if err := result.FromSecurityExceptionsAPIExceptionListItemEntryList(apiEntry); err != nil {
			diags.AddError("Failed to create list entry", err.Error())
		}

	case "exists":
		apiEntry := kbapi.SecurityExceptionsAPIExceptionListItemEntryExists{
			Type:     "exists",
			Field:    field,
			Operator: operator,
		}
		if err := result.FromSecurityExceptionsAPIExceptionListItemEntryExists(apiEntry); err != nil {
			diags.AddError("Failed to create exists entry", err.Error())
		}

	case "wildcard":
		// Validate required field
		if entry.Value.IsNull() || entry.Value.IsUnknown() || entry.Value.ValueString() == "" {
			diags.AddError("Invalid Configuration", "Attribute 'value' is required when type is 'wildcard'")
			return result, diags
		}

		apiEntry := kbapi.SecurityExceptionsAPIExceptionListItemEntryMatchWildcard{
			Type:     "wildcard",
			Field:    field,
			Operator: operator,
			Value:    kbapi.SecurityExceptionsAPINonEmptyString(entry.Value.ValueString()),
		}
		if err := result.FromSecurityExceptionsAPIExceptionListItemEntryMatchWildcard(apiEntry); err != nil {
			diags.AddError("Failed to create wildcard entry", err.Error())
		}

	case "nested":
		// Validate required field
		if entry.Entries.IsNull() || entry.Entries.IsUnknown() {
			diags.AddError("Invalid Configuration", "Attribute 'entries' is required when type is 'nested'")
			return result, diags
		}

		var nestedEntries []NestedEntryModel
		diags.Append(entry.Entries.ElementsAs(ctx, &nestedEntries, false)...)
		if diags.HasError() {
			return result, diags
		}

		if len(nestedEntries) == 0 {
			diags.AddError("Invalid Configuration", "Attribute 'entries' must contain at least one entry when type is 'nested'")
			return result, diags
		}

		apiNestedEntries := make([]kbapi.SecurityExceptionsAPIExceptionListItemEntryNestedEntryItem, 0, len(nestedEntries))
		for _, ne := range nestedEntries {
			nestedAPIEntry, d := convertNestedEntryToAPI(ctx, ne)
			diags.Append(d...)
			if d.HasError() {
				continue
			}
			apiNestedEntries = append(apiNestedEntries, nestedAPIEntry)
		}

		apiEntry := kbapi.SecurityExceptionsAPIExceptionListItemEntryNested{
			Type:    "nested",
			Field:   field,
			Entries: apiNestedEntries,
		}
		if err := result.FromSecurityExceptionsAPIExceptionListItemEntryNested(apiEntry); err != nil {
			diags.AddError("Failed to create nested entry", err.Error())
		}

	default:
		diags.AddError("Invalid entry type", fmt.Sprintf("Unknown entry type: %s", entryType))
	}

	return result, diags
}

// convertNestedEntryToAPI converts a nested entry model to an API nested entry model
func convertNestedEntryToAPI(ctx context.Context, entry NestedEntryModel) (kbapi.SecurityExceptionsAPIExceptionListItemEntryNestedEntryItem, diag.Diagnostics) {
	var diags diag.Diagnostics
	var result kbapi.SecurityExceptionsAPIExceptionListItemEntryNestedEntryItem

	entryType := entry.Type.ValueString()
	operator := kbapi.SecurityExceptionsAPIExceptionListItemEntryOperator(entry.Operator.ValueString())
	field := kbapi.SecurityExceptionsAPINonEmptyString(entry.Field.ValueString())

	switch entryType {
	case "match":
		// Validate required field
		if entry.Value.IsNull() || entry.Value.IsUnknown() || entry.Value.ValueString() == "" {
			diags.AddError("Invalid Configuration", "Attribute 'value' is required for nested entry when type is 'match'")
			return result, diags
		}

		apiEntry := kbapi.SecurityExceptionsAPIExceptionListItemEntryMatch{
			Type:     "match",
			Field:    field,
			Operator: operator,
			Value:    kbapi.SecurityExceptionsAPINonEmptyString(entry.Value.ValueString()),
		}
		if err := result.FromSecurityExceptionsAPIExceptionListItemEntryMatch(apiEntry); err != nil {
			diags.AddError("Failed to create nested match entry", err.Error())
		}

	case "match_any":
		// Validate required field
		if entry.Values.IsNull() || entry.Values.IsUnknown() {
			diags.AddError("Invalid Configuration", "Attribute 'values' is required for nested entry when type is 'match_any'")
			return result, diags
		}

		var values []string
		diags.Append(entry.Values.ElementsAs(ctx, &values, false)...)
		if diags.HasError() {
			return result, diags
		}

		if len(values) == 0 {
			diags.AddError("Invalid Configuration", "Attribute 'values' must contain at least one value for nested entry when type is 'match_any'")
			return result, diags
		}

		apiValues := make([]kbapi.SecurityExceptionsAPINonEmptyString, len(values))
		for i, v := range values {
			apiValues[i] = kbapi.SecurityExceptionsAPINonEmptyString(v)
		}
		apiEntry := kbapi.SecurityExceptionsAPIExceptionListItemEntryMatchAny{
			Type:     "match_any",
			Field:    field,
			Operator: operator,
			Value:    apiValues,
		}
		if err := result.FromSecurityExceptionsAPIExceptionListItemEntryMatchAny(apiEntry); err != nil {
			diags.AddError("Failed to create nested match_any entry", err.Error())
		}

	case "exists":
		apiEntry := kbapi.SecurityExceptionsAPIExceptionListItemEntryExists{
			Type:     "exists",
			Field:    field,
			Operator: operator,
		}
		if err := result.FromSecurityExceptionsAPIExceptionListItemEntryExists(apiEntry); err != nil {
			diags.AddError("Failed to create nested exists entry", err.Error())
		}

	default:
		diags.AddError("Invalid nested entry type", fmt.Sprintf("Unknown nested entry type: %s. Only 'match', 'match_any', and 'exists' are allowed.", entryType))
	}

	return result, diags
}

// convertEntriesFromAPI converts API entry models to Terraform entry models
func convertEntriesFromAPI(ctx context.Context, apiEntries kbapi.SecurityExceptionsAPIExceptionListItemEntryArray) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics

	if len(apiEntries) == 0 {
		return types.ListNull(types.ObjectType{
			AttrTypes: getEntryAttrTypes(),
		}), diags
	}

	entries := make([]EntryModel, 0, len(apiEntries))
	for _, apiEntry := range apiEntries {
		entry, d := convertEntryFromAPI(ctx, apiEntry)
		diags.Append(d...)
		if d.HasError() {
			continue
		}
		entries = append(entries, entry)
	}

	list, d := types.ListValueFrom(ctx, types.ObjectType{
		AttrTypes: getEntryAttrTypes(),
	}, entries)
	diags.Append(d...)
	return list, diags
}

// convertEntryFromAPI converts a single API entry to a Terraform entry model
func convertEntryFromAPI(ctx context.Context, apiEntry kbapi.SecurityExceptionsAPIExceptionListItemEntry) (EntryModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	var entry EntryModel

	// Marshal the entry back to JSON to inspect its type
	entryBytes, err := apiEntry.MarshalJSON()
	if err != nil {
		diags.AddError("Failed to marshal entry", err.Error())
		return entry, diags
	}

	// Try to unmarshal into a map to determine the type
	var entryMap map[string]interface{}
	if err := json.Unmarshal(entryBytes, &entryMap); err != nil {
		diags.AddError("Failed to unmarshal entry", err.Error())
		return entry, diags
	}

	entryType, ok := entryMap["type"].(string)
	if !ok {
		diags.AddError("Invalid entry", "Entry is missing 'type' field")
		return entry, diags
	}

	entry.Type = types.StringValue(entryType)
	if field, ok := entryMap["field"].(string); ok {
		entry.Field = types.StringValue(field)
	}
	if operator, ok := entryMap["operator"].(string); ok {
		entry.Operator = types.StringValue(operator)
	}

	switch entryType {
	case "match", "wildcard":
		if value, ok := entryMap["value"].(string); ok {
			entry.Value = types.StringValue(value)
		} else {
			entry.Value = types.StringNull()
		}
		entry.Values = types.ListNull(types.StringType)
		entry.List = types.ObjectNull(getListAttrTypes())
		entry.Entries = types.ListNull(types.ObjectType{AttrTypes: getNestedEntryAttrTypes()})

	case "match_any":
		if values, ok := entryMap["value"].([]interface{}); ok {
			strValues := make([]string, 0, len(values))
			for _, v := range values {
				if str, ok := v.(string); ok {
					strValues = append(strValues, str)
				}
			}
			list, d := types.ListValueFrom(ctx, types.StringType, strValues)
			diags.Append(d...)
			entry.Values = list
		} else {
			entry.Values = types.ListNull(types.StringType)
		}
		entry.Value = types.StringNull()
		entry.List = types.ObjectNull(getListAttrTypes())
		entry.Entries = types.ListNull(types.ObjectType{AttrTypes: getNestedEntryAttrTypes()})

	case "list":
		if listData, ok := entryMap["list"].(map[string]interface{}); ok {
			listModel := EntryListModel{
				ID:   types.StringValue(listData["id"].(string)),
				Type: types.StringValue(listData["type"].(string)),
			}
			obj, d := types.ObjectValueFrom(ctx, getListAttrTypes(), listModel)
			diags.Append(d...)
			entry.List = obj
		} else {
			entry.List = types.ObjectNull(getListAttrTypes())
		}
		entry.Value = types.StringNull()
		entry.Values = types.ListNull(types.StringType)
		entry.Entries = types.ListNull(types.ObjectType{AttrTypes: getNestedEntryAttrTypes()})

	case "exists":
		entry.Value = types.StringNull()
		entry.Values = types.ListNull(types.StringType)
		entry.List = types.ObjectNull(getListAttrTypes())
		entry.Entries = types.ListNull(types.ObjectType{AttrTypes: getNestedEntryAttrTypes()})

	case "nested":
		// Nested entries don't have an operator field in the API
		entry.Operator = types.StringNull()
		if entriesData, ok := entryMap["entries"].([]interface{}); ok {
			nestedEntries := make([]NestedEntryModel, 0, len(entriesData))
			for _, neData := range entriesData {
				if neMap, ok := neData.(map[string]interface{}); ok {
					ne, d := convertNestedEntryFromMap(ctx, neMap)
					diags.Append(d...)
					if !d.HasError() {
						nestedEntries = append(nestedEntries, ne)
					}
				}
			}
			list, d := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: getNestedEntryAttrTypes()}, nestedEntries)
			diags.Append(d...)
			entry.Entries = list
		} else {
			entry.Entries = types.ListNull(types.ObjectType{AttrTypes: getNestedEntryAttrTypes()})
		}
		entry.Value = types.StringNull()
		entry.Values = types.ListNull(types.StringType)
		entry.List = types.ObjectNull(getListAttrTypes())
	}

	return entry, diags
}

// convertNestedEntryFromMap converts a map representation of nested entry to a model
func convertNestedEntryFromMap(ctx context.Context, entryMap map[string]interface{}) (NestedEntryModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	var entry NestedEntryModel

	if entryType, ok := entryMap["type"].(string); ok {
		entry.Type = types.StringValue(entryType)
	}
	if field, ok := entryMap["field"].(string); ok {
		entry.Field = types.StringValue(field)
	}
	if operator, ok := entryMap["operator"].(string); ok {
		entry.Operator = types.StringValue(operator)
	}

	entryType := entry.Type.ValueString()
	switch entryType {
	case "match":
		if value, ok := entryMap["value"].(string); ok {
			entry.Value = types.StringValue(value)
		} else {
			entry.Value = types.StringNull()
		}
		entry.Values = types.ListNull(types.StringType)

	case "match_any":
		if values, ok := entryMap["value"].([]interface{}); ok {
			strValues := make([]string, 0, len(values))
			for _, v := range values {
				if str, ok := v.(string); ok {
					strValues = append(strValues, str)
				}
			}
			list, d := types.ListValueFrom(ctx, types.StringType, strValues)
			diags.Append(d...)
			entry.Values = list
		} else {
			entry.Values = types.ListNull(types.StringType)
		}
		entry.Value = types.StringNull()

	case "exists":
		entry.Value = types.StringNull()
		entry.Values = types.ListNull(types.StringType)
	}

	return entry, diags
}

// getEntryAttrTypes returns the attribute types for entry objects
func getEntryAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"type":     types.StringType,
		"field":    types.StringType,
		"operator": types.StringType,
		"value":    types.StringType,
		"values":   types.ListType{ElemType: types.StringType},
		"list":     types.ObjectType{AttrTypes: getListAttrTypes()},
		"entries":  types.ListType{ElemType: types.ObjectType{AttrTypes: getNestedEntryAttrTypes()}},
	}
}

// getListAttrTypes returns the attribute types for list objects
func getListAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":   types.StringType,
		"type": types.StringType,
	}
}

// getNestedEntryAttrTypes returns the attribute types for nested entry objects
func getNestedEntryAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"type":     types.StringType,
		"field":    types.StringType,
		"operator": types.StringType,
		"value":    types.StringType,
		"values":   types.ListType{ElemType: types.StringType},
	}
}
