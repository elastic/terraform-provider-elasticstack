package security_exception_item

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// MinVersionExpireTime defines the minimum server version required for expire_time field
var MinVersionExpireTime = version.Must(version.NewVersion("8.7.2"))

type ExceptionItemModel struct {
	ID            types.String         `tfsdk:"id"`
	SpaceID       types.String         `tfsdk:"space_id"`
	ItemID        types.String         `tfsdk:"item_id"`
	ListID        types.String         `tfsdk:"list_id"`
	Name          types.String         `tfsdk:"name"`
	Description   types.String         `tfsdk:"description"`
	Type          types.String         `tfsdk:"type"`
	NamespaceType types.String         `tfsdk:"namespace_type"`
	OsTypes       types.Set            `tfsdk:"os_types"`
	Tags          types.Set            `tfsdk:"tags"`
	Meta          jsontypes.Normalized `tfsdk:"meta"`
	Entries       types.List           `tfsdk:"entries"`
	Comments      types.List           `tfsdk:"comments"`
	ExpireTime    timetypes.RFC3339    `tfsdk:"expire_time"`
	CreatedAt     types.String         `tfsdk:"created_at"`
	CreatedBy     types.String         `tfsdk:"created_by"`
	UpdatedAt     types.String         `tfsdk:"updated_at"`
	UpdatedBy     types.String         `tfsdk:"updated_by"`
	TieBreakerID  types.String         `tfsdk:"tie_breaker_id"`
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

	if !utils.IsKnown(entries) {
		return nil, diags
	}

	entryModels := utils.ListTypeAs[EntryModel](ctx, entries, path.Empty(), &diags)
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

// convertMatchEntryToAPI converts a match entry to API format
func convertMatchEntryToAPI(entry EntryModel, field kbapi.SecurityExceptionsAPINonEmptyString, operator kbapi.SecurityExceptionsAPIExceptionListItemEntryOperator) (kbapi.SecurityExceptionsAPIExceptionListItemEntry, diag.Diagnostics) {
	var diags diag.Diagnostics
	var result kbapi.SecurityExceptionsAPIExceptionListItemEntry

	// Validate required field
	if !utils.IsKnown(entry.Value) || entry.Value.ValueString() == "" {
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

	return result, diags
}

// convertMatchAnyEntryToAPI converts a match_any entry to API format
func convertMatchAnyEntryToAPI(ctx context.Context, entry EntryModel, field kbapi.SecurityExceptionsAPINonEmptyString, operator kbapi.SecurityExceptionsAPIExceptionListItemEntryOperator) (kbapi.SecurityExceptionsAPIExceptionListItemEntry, diag.Diagnostics) {
	var diags diag.Diagnostics
	var result kbapi.SecurityExceptionsAPIExceptionListItemEntry

	// Validate required field
	if !utils.IsKnown(entry.Values) {
		diags.AddError("Invalid Configuration", "Attribute 'values' is required when type is 'match_any'")
		return result, diags
	}

	values := utils.ListTypeAs[string](ctx, entry.Values, path.Empty(), &diags)
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

	return result, diags
}

// convertListEntryToAPI converts a list entry to API format
func convertListEntryToAPI(ctx context.Context, entry EntryModel, field kbapi.SecurityExceptionsAPINonEmptyString, operator kbapi.SecurityExceptionsAPIExceptionListItemEntryOperator) (kbapi.SecurityExceptionsAPIExceptionListItemEntry, diag.Diagnostics) {
	var diags diag.Diagnostics
	var result kbapi.SecurityExceptionsAPIExceptionListItemEntry

	// Validate required field
	if !utils.IsKnown(entry.List) {
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

	return result, diags
}

// convertExistsEntryToAPI converts an exists entry to API format
func convertExistsEntryToAPI(field kbapi.SecurityExceptionsAPINonEmptyString, operator kbapi.SecurityExceptionsAPIExceptionListItemEntryOperator) (kbapi.SecurityExceptionsAPIExceptionListItemEntry, diag.Diagnostics) {
	var diags diag.Diagnostics
	var result kbapi.SecurityExceptionsAPIExceptionListItemEntry

	apiEntry := kbapi.SecurityExceptionsAPIExceptionListItemEntryExists{
		Type:     "exists",
		Field:    field,
		Operator: operator,
	}
	if err := result.FromSecurityExceptionsAPIExceptionListItemEntryExists(apiEntry); err != nil {
		diags.AddError("Failed to create exists entry", err.Error())
	}

	return result, diags
}

// convertWildcardEntryToAPI converts a wildcard entry to API format
func convertWildcardEntryToAPI(entry EntryModel, field kbapi.SecurityExceptionsAPINonEmptyString, operator kbapi.SecurityExceptionsAPIExceptionListItemEntryOperator) (kbapi.SecurityExceptionsAPIExceptionListItemEntry, diag.Diagnostics) {
	var diags diag.Diagnostics
	var result kbapi.SecurityExceptionsAPIExceptionListItemEntry

	// Validate required field
	if !utils.IsKnown(entry.Value) || entry.Value.ValueString() == "" {
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

	return result, diags
}

// convertNestedEntryArrayToAPI converts nested entries to API format
func convertNestedEntryArrayToAPI(ctx context.Context, entry EntryModel, field kbapi.SecurityExceptionsAPINonEmptyString) (kbapi.SecurityExceptionsAPIExceptionListItemEntry, diag.Diagnostics) {
	var diags diag.Diagnostics
	var result kbapi.SecurityExceptionsAPIExceptionListItemEntry

	// Validate required field
	if !utils.IsKnown(entry.Entries) {
		diags.AddError("Invalid Configuration", "Attribute 'entries' is required when type is 'nested'")
		return result, diags
	}

	nestedEntries := utils.ListTypeAs[NestedEntryModel](ctx, entry.Entries, path.Empty(), &diags)
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

	return result, diags
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
		return convertMatchEntryToAPI(entry, field, operator)
	case "match_any":
		return convertMatchAnyEntryToAPI(ctx, entry, field, operator)
	case "list":
		return convertListEntryToAPI(ctx, entry, field, operator)
	case "exists":
		return convertExistsEntryToAPI(field, operator)
	case "wildcard":
		return convertWildcardEntryToAPI(entry, field, operator)
	case "nested":
		return convertNestedEntryArrayToAPI(ctx, entry, field)
	default:
		diags.AddError("Invalid entry type", fmt.Sprintf("Unknown entry type: %s", entryType))
		return result, diags
	}
}

// convertNestedMatchEntryToAPI converts a nested match entry to API format
func convertNestedMatchEntryToAPI(entry NestedEntryModel, field kbapi.SecurityExceptionsAPINonEmptyString, operator kbapi.SecurityExceptionsAPIExceptionListItemEntryOperator) (kbapi.SecurityExceptionsAPIExceptionListItemEntryNestedEntryItem, diag.Diagnostics) {
	var diags diag.Diagnostics
	var result kbapi.SecurityExceptionsAPIExceptionListItemEntryNestedEntryItem

	// Validate required field
	if !utils.IsKnown(entry.Value) || entry.Value.ValueString() == "" {
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

	return result, diags
}

// convertNestedMatchAnyEntryToAPI converts a nested match_any entry to API format
func convertNestedMatchAnyEntryToAPI(ctx context.Context, entry NestedEntryModel, field kbapi.SecurityExceptionsAPINonEmptyString, operator kbapi.SecurityExceptionsAPIExceptionListItemEntryOperator) (kbapi.SecurityExceptionsAPIExceptionListItemEntryNestedEntryItem, diag.Diagnostics) {
	var diags diag.Diagnostics
	var result kbapi.SecurityExceptionsAPIExceptionListItemEntryNestedEntryItem

	// Validate required field
	if !utils.IsKnown(entry.Values) {
		diags.AddError("Invalid Configuration", "Attribute 'values' is required for nested entry when type is 'match_any'")
		return result, diags
	}

	values := utils.ListTypeAs[string](ctx, entry.Values, path.Empty(), &diags)
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

	return result, diags
}

// convertNestedExistsEntryToAPI converts a nested exists entry to API format
func convertNestedExistsEntryToAPI(field kbapi.SecurityExceptionsAPINonEmptyString, operator kbapi.SecurityExceptionsAPIExceptionListItemEntryOperator) (kbapi.SecurityExceptionsAPIExceptionListItemEntryNestedEntryItem, diag.Diagnostics) {
	var diags diag.Diagnostics
	var result kbapi.SecurityExceptionsAPIExceptionListItemEntryNestedEntryItem

	apiEntry := kbapi.SecurityExceptionsAPIExceptionListItemEntryExists{
		Type:     "exists",
		Field:    field,
		Operator: operator,
	}
	if err := result.FromSecurityExceptionsAPIExceptionListItemEntryExists(apiEntry); err != nil {
		diags.AddError("Failed to create nested exists entry", err.Error())
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
		return convertNestedMatchEntryToAPI(entry, field, operator)
	case "match_any":
		return convertNestedMatchAnyEntryToAPI(ctx, entry, field, operator)
	case "exists":
		return convertNestedExistsEntryToAPI(field, operator)
	default:
		diags.AddError("Invalid nested entry type", fmt.Sprintf("Unknown nested entry type: %s. Only 'match', 'match_any', and 'exists' are allowed.", entryType))
		return result, diags
	}
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

// convertMatchOrWildcardEntryFromAPI converts match or wildcard entries from API format
func convertMatchOrWildcardEntryFromAPI(entryMap map[string]interface{}, entry *EntryModel) {
	if value, ok := entryMap["value"].(string); ok {
		entry.Value = types.StringValue(value)
	} else {
		entry.Value = types.StringNull()
	}
	entry.Values = types.ListNull(types.StringType)
	entry.List = types.ObjectNull(getListAttrTypes())
	entry.Entries = types.ListNull(types.ObjectType{AttrTypes: getNestedEntryAttrTypes()})
}

// convertMatchAnyEntryFromAPI converts match_any entries from API format
func convertMatchAnyEntryFromAPI(ctx context.Context, entryMap map[string]interface{}, entry *EntryModel) diag.Diagnostics {
	var diags diag.Diagnostics

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
	return diags
}

// convertListEntryFromAPI converts list entries from API format
func convertListEntryFromAPI(ctx context.Context, entryMap map[string]interface{}, entry *EntryModel) diag.Diagnostics {
	var diags diag.Diagnostics

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
	return diags
}

// convertExistsEntryFromAPI converts exists entries from API format
func convertExistsEntryFromAPI(entry *EntryModel) {
	entry.Value = types.StringNull()
	entry.Values = types.ListNull(types.StringType)
	entry.List = types.ObjectNull(getListAttrTypes())
	entry.Entries = types.ListNull(types.ObjectType{AttrTypes: getNestedEntryAttrTypes()})
}

// convertNestedEntryFromAPI converts nested entries from API format
func convertNestedEntryFromAPI(ctx context.Context, entryMap map[string]interface{}, entry *EntryModel) diag.Diagnostics {
	var diags diag.Diagnostics

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
	return diags
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
		convertMatchOrWildcardEntryFromAPI(entryMap, &entry)
	case "match_any":
		d := convertMatchAnyEntryFromAPI(ctx, entryMap, &entry)
		diags.Append(d...)
	case "list":
		d := convertListEntryFromAPI(ctx, entryMap, &entry)
		diags.Append(d...)
	case "exists":
		convertExistsEntryFromAPI(&entry)
	case "nested":
		d := convertNestedEntryFromAPI(ctx, entryMap, &entry)
		diags.Append(d...)
	}

	return entry, diags
}

// convertNestedMatchFromMap converts nested match entries from map format
func convertNestedMatchFromMap(entryMap map[string]interface{}, entry *NestedEntryModel) {
	if value, ok := entryMap["value"].(string); ok {
		entry.Value = types.StringValue(value)
	} else {
		entry.Value = types.StringNull()
	}
	entry.Values = types.ListNull(types.StringType)
}

// convertNestedMatchAnyFromMap converts nested match_any entries from map format
func convertNestedMatchAnyFromMap(ctx context.Context, entryMap map[string]interface{}, entry *NestedEntryModel) diag.Diagnostics {
	var diags diag.Diagnostics

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
	return diags
}

// convertNestedExistsFromMap converts nested exists entries from map format
func convertNestedExistsFromMap(entry *NestedEntryModel) {
	entry.Value = types.StringNull()
	entry.Values = types.ListNull(types.StringType)
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
		convertNestedMatchFromMap(entryMap, &entry)
	case "match_any":
		d := convertNestedMatchAnyFromMap(ctx, entryMap, &entry)
		diags.Append(d...)
	case "exists":
		convertNestedExistsFromMap(&entry)
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

// getCommentAttrTypes returns the attribute types for comment objects
func getCommentAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":      types.StringType,
		"comment": types.StringType,
	}
}

// convertEntriesToAPIWithDiags converts entries and handles diagnostics
func (m *ExceptionItemModel) convertEntriesToAPIWithDiags(ctx context.Context, diags *diag.Diagnostics) kbapi.SecurityExceptionsAPIExceptionListItemEntryArray {
	entries, d := convertEntriesToAPI(ctx, m.Entries)
	diags.Append(d...)
	return entries
}

// CommonExceptionItemProps holds pointers to common fields across create/update requests
type CommonExceptionItemProps struct {
	NamespaceType *kbapi.SecurityExceptionsAPIExceptionNamespaceType
	OsTypes       *[]kbapi.SecurityExceptionsAPIExceptionListOsType
	Tags          *kbapi.SecurityExceptionsAPIExceptionListItemTags
	Meta          *kbapi.SecurityExceptionsAPIExceptionListItemMeta
	ExpireTime    *kbapi.SecurityExceptionsAPIExceptionListItemExpireTime
}

// setCommonProps sets common fields across create and update requests
func (m *ExceptionItemModel) setCommonProps(
	ctx context.Context,
	props *CommonExceptionItemProps,
	diags *diag.Diagnostics,
	client clients.MinVersionEnforceable,
) {
	// Set optional namespace_type
	if utils.IsKnown(m.NamespaceType) {
		nsType := kbapi.SecurityExceptionsAPIExceptionNamespaceType(m.NamespaceType.ValueString())
		*props.NamespaceType = nsType
	}

	// Set optional os_types
	if utils.IsKnown(m.OsTypes) {
		osTypes := utils.SetTypeAs[kbapi.SecurityExceptionsAPIExceptionListOsType](ctx, m.OsTypes, path.Empty(), diags)
		if diags.HasError() {
			return
		}
		if len(osTypes) > 0 {
			*props.OsTypes = osTypes
		}
	}

	// Set optional tags
	if utils.IsKnown(m.Tags) {
		tags := utils.SetTypeAs[string](ctx, m.Tags, path.Empty(), diags)
		if diags.HasError() {
			return
		}
		if len(tags) > 0 {
			tagsArray := kbapi.SecurityExceptionsAPIExceptionListItemTags(tags)
			*props.Tags = tagsArray
		}
	}

	// Set optional meta
	if utils.IsKnown(m.Meta) {
		var meta kbapi.SecurityExceptionsAPIExceptionListItemMeta
		unmarshalDiags := m.Meta.Unmarshal(&meta)
		diags.Append(unmarshalDiags...)
		if diags.HasError() {
			return
		}
		*props.Meta = meta
	}

	// Set optional expire_time
	if utils.IsKnown(m.ExpireTime) {
		// Check version support for expire_time
		if supported, versionDiags := client.EnforceMinVersion(ctx, MinVersionExpireTime); versionDiags.HasError() {
			diags.Append(diagutil.FrameworkDiagsFromSDK(versionDiags)...)
			return
		} else if !supported {
			diags.AddError("expire_time is unsupported",
				fmt.Sprintf("expire_time requires server version %s or higher", MinVersionExpireTime.String()))
			return
		}

		expireTime, d := m.ExpireTime.ValueRFC3339Time()
		diags.Append(d...)
		if diags.HasError() {
			return
		}

		expireTimeAPI := kbapi.SecurityExceptionsAPIExceptionListItemExpireTime(expireTime.Format("2006-01-02T15:04:05.000Z"))
		*props.ExpireTime = expireTimeAPI
	}
}

// commentsToCreateAPI converts comments to create API format
func (m *ExceptionItemModel) commentsToCreateAPI(
	ctx context.Context,
	diags *diag.Diagnostics,
) *kbapi.SecurityExceptionsAPICreateExceptionListItemCommentArray {
	if !utils.IsKnown(m.Comments) {
		return nil
	}

	comments := utils.ListTypeAs[CommentModel](ctx, m.Comments, path.Empty(), diags)
	if diags.HasError() || len(comments) == 0 {
		return nil
	}

	commentsArray := make(kbapi.SecurityExceptionsAPICreateExceptionListItemCommentArray, len(comments))
	for i, comment := range comments {
		commentsArray[i] = kbapi.SecurityExceptionsAPICreateExceptionListItemComment{
			Comment: kbapi.SecurityExceptionsAPINonEmptyString(comment.Comment.ValueString()),
		}
	}
	return &commentsArray
}

// commentsToUpdateAPI converts comments to update API format
func (m *ExceptionItemModel) commentsToUpdateAPI(
	ctx context.Context,
	diags *diag.Diagnostics,
) *kbapi.SecurityExceptionsAPIUpdateExceptionListItemCommentArray {
	if !utils.IsKnown(m.Comments) {
		return nil
	}

	comments := utils.ListTypeAs[CommentModel](ctx, m.Comments, path.Empty(), diags)
	if diags.HasError() || len(comments) == 0 {
		return nil
	}

	commentsArray := make(kbapi.SecurityExceptionsAPIUpdateExceptionListItemCommentArray, len(comments))
	for i, comment := range comments {
		commentsArray[i] = kbapi.SecurityExceptionsAPIUpdateExceptionListItemComment{
			Comment: kbapi.SecurityExceptionsAPINonEmptyString(comment.Comment.ValueString()),
		}
	}
	return &commentsArray
}

// toCreateRequest converts the Terraform model to API create request
func (m *ExceptionItemModel) toCreateRequest(ctx context.Context, client clients.MinVersionEnforceable) (*kbapi.CreateExceptionListItemJSONRequestBody, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Convert entries from Terraform model to API model
	entries := m.convertEntriesToAPIWithDiags(ctx, &diags)
	if diags.HasError() {
		return nil, diags
	}

	req := &kbapi.CreateExceptionListItemJSONRequestBody{
		ListId:      kbapi.SecurityExceptionsAPIExceptionListHumanId(m.ListID.ValueString()),
		Name:        kbapi.SecurityExceptionsAPIExceptionListItemName(m.Name.ValueString()),
		Description: kbapi.SecurityExceptionsAPIExceptionListItemDescription(m.Description.ValueString()),
		Type:        kbapi.SecurityExceptionsAPIExceptionListItemType(m.Type.ValueString()),
		Entries:     entries,
	}

	// Set optional item_id
	if utils.IsKnown(m.ItemID) {
		itemID := kbapi.SecurityExceptionsAPIExceptionListItemHumanId(m.ItemID.ValueString())
		req.ItemId = &itemID
	}

	// Set common properties
	var nsType kbapi.SecurityExceptionsAPIExceptionNamespaceType
	var osTypes []kbapi.SecurityExceptionsAPIExceptionListOsType
	var tags kbapi.SecurityExceptionsAPIExceptionListItemTags
	var meta kbapi.SecurityExceptionsAPIExceptionListItemMeta
	var expireTime kbapi.SecurityExceptionsAPIExceptionListItemExpireTime

	m.setCommonProps(ctx, &CommonExceptionItemProps{
		NamespaceType: &nsType,
		OsTypes:       &osTypes,
		Tags:          &tags,
		Meta:          &meta,
		ExpireTime:    &expireTime,
	}, &diags, client)
	if diags.HasError() {
		return nil, diags
	}

	// Assign common properties to request if they were set
	if utils.IsKnown(m.NamespaceType) {
		req.NamespaceType = &nsType
	}
	if utils.IsKnown(m.OsTypes) && len(osTypes) > 0 {
		req.OsTypes = &osTypes
	}
	if utils.IsKnown(m.Tags) && len(tags) > 0 {
		req.Tags = &tags
	}
	if utils.IsKnown(m.Meta) {
		req.Meta = &meta
	}
	if utils.IsKnown(m.ExpireTime) {
		req.ExpireTime = &expireTime
	}

	// Set optional comments
	if comments := m.commentsToCreateAPI(ctx, &diags); comments != nil {
		req.Comments = comments
	}
	if diags.HasError() {
		return nil, diags
	}

	return req, diags
}

// toUpdateRequest converts the Terraform model to API update request
func (m *ExceptionItemModel) toUpdateRequest(ctx context.Context, resourceId string, client clients.MinVersionEnforceable) (*kbapi.UpdateExceptionListItemJSONRequestBody, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Convert entries from Terraform model to API model
	entries := m.convertEntriesToAPIWithDiags(ctx, &diags)
	if diags.HasError() {
		return nil, diags
	}

	id := kbapi.SecurityExceptionsAPIExceptionListItemId(resourceId)
	req := &kbapi.UpdateExceptionListItemJSONRequestBody{
		Id:          &id,
		Name:        kbapi.SecurityExceptionsAPIExceptionListItemName(m.Name.ValueString()),
		Description: kbapi.SecurityExceptionsAPIExceptionListItemDescription(m.Description.ValueString()),
		Type:        kbapi.SecurityExceptionsAPIExceptionListItemType(m.Type.ValueString()),
		Entries:     entries,
	}

	// Set common properties
	var nsType kbapi.SecurityExceptionsAPIExceptionNamespaceType
	var osTypes []kbapi.SecurityExceptionsAPIExceptionListOsType
	var tags kbapi.SecurityExceptionsAPIExceptionListItemTags
	var meta kbapi.SecurityExceptionsAPIExceptionListItemMeta
	var expireTime kbapi.SecurityExceptionsAPIExceptionListItemExpireTime

	m.setCommonProps(ctx, &CommonExceptionItemProps{
		NamespaceType: &nsType,
		OsTypes:       &osTypes,
		Tags:          &tags,
		Meta:          &meta,
		ExpireTime:    &expireTime,
	}, &diags, client)
	if diags.HasError() {
		return nil, diags
	}

	// Assign common properties to request if they were set
	if utils.IsKnown(m.NamespaceType) {
		req.NamespaceType = &nsType
	}
	if utils.IsKnown(m.OsTypes) && len(osTypes) > 0 {
		req.OsTypes = &osTypes
	}
	if utils.IsKnown(m.Tags) && len(tags) > 0 {
		req.Tags = &tags
	}
	if utils.IsKnown(m.Meta) {
		req.Meta = &meta
	}
	if utils.IsKnown(m.ExpireTime) {
		req.ExpireTime = &expireTime
	}

	// Set optional comments
	if comments := m.commentsToUpdateAPI(ctx, &diags); comments != nil {
		req.Comments = comments
	}
	if diags.HasError() {
		return nil, diags
	}

	return req, diags
}

// fromAPI converts the API response to Terraform model
func (m *ExceptionItemModel) fromAPI(ctx context.Context, apiResp *kbapi.SecurityExceptionsAPIExceptionListItem) diag.Diagnostics {
	var diags diag.Diagnostics

	// Create composite ID from space_id and item id
	compId := clients.CompositeId{
		ClusterId:  m.SpaceID.ValueString(),
		ResourceId: typeutils.StringishValue(apiResp.Id).ValueString(),
	}
	m.ID = types.StringValue(compId.String())

	m.ItemID = typeutils.StringishValue(apiResp.ItemId)
	m.ListID = typeutils.StringishValue(apiResp.ListId)
	m.Name = typeutils.StringishValue(apiResp.Name)
	m.Description = typeutils.StringishValue(apiResp.Description)
	m.Type = typeutils.StringishValue(apiResp.Type)
	m.NamespaceType = typeutils.StringishValue(apiResp.NamespaceType)
	m.CreatedAt = types.StringValue(apiResp.CreatedAt.Format("2006-01-02T15:04:05.000Z"))
	m.CreatedBy = types.StringValue(apiResp.CreatedBy)
	m.UpdatedAt = types.StringValue(apiResp.UpdatedAt.Format("2006-01-02T15:04:05.000Z"))
	m.UpdatedBy = types.StringValue(apiResp.UpdatedBy)
	m.TieBreakerID = types.StringValue(apiResp.TieBreakerId)

	// Set optional expire_time
	if apiResp.ExpireTime != nil {
		expireTime, err := time.Parse(time.RFC3339, string(*apiResp.ExpireTime))
		if err != nil {
			diags.AddError("Failed to parse expire_time from API response", err.Error())
			m.ExpireTime = timetypes.NewRFC3339Null()
		} else {
			m.ExpireTime = timetypes.NewRFC3339TimeValue(expireTime)
		}
	} else {
		m.ExpireTime = timetypes.NewRFC3339Null()
	}

	// Set optional os_types
	if apiResp.OsTypes != nil && len(*apiResp.OsTypes) > 0 {
		set, d := types.SetValueFrom(ctx, types.StringType, *apiResp.OsTypes)
		diags.Append(d...)
		m.OsTypes = set
	} else {
		m.OsTypes = types.SetNull(types.StringType)
	}

	// Set optional tags
	if apiResp.Tags != nil && len(*apiResp.Tags) > 0 {
		set, d := types.SetValueFrom(ctx, types.StringType, *apiResp.Tags)
		diags.Append(d...)
		m.Tags = set
	} else {
		m.Tags = types.SetNull(types.StringType)
	}

	// Set optional meta
	if apiResp.Meta != nil {
		metaBytes, err := json.Marshal(apiResp.Meta)
		if err != nil {
			diags.AddError("Failed to marshal meta field from API response to JSON", err.Error())
			return diags
		}
		m.Meta = jsontypes.NewNormalizedValue(string(metaBytes))
	} else {
		m.Meta = jsontypes.NewNormalizedNull()
	}

	// Set entries (convert from API model to Terraform model)
	entriesList, d := convertEntriesFromAPI(ctx, apiResp.Entries)
	diags.Append(d...)
	m.Entries = entriesList

	// Set optional comments
	if len(apiResp.Comments) > 0 {
		comments := make([]CommentModel, len(apiResp.Comments))
		for i, comment := range apiResp.Comments {
			comments[i] = CommentModel{
				ID:      typeutils.StringishValue(comment.Id),
				Comment: typeutils.StringishValue(comment.Comment),
			}
		}
		list, d := types.ListValueFrom(ctx, types.ObjectType{
			AttrTypes: getCommentAttrTypes(),
		}, comments)
		diags.Append(d...)
		m.Comments = list
	} else {
		m.Comments = types.ListNull(types.ObjectType{
			AttrTypes: getCommentAttrTypes(),
		})
	}

	return diags
}
