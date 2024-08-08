package indices

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Read refreshes the Terraform state with the latest data.
func (d *dataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state dataSourceModel

	// Resolve search attribute
	var search string
	diag := req.Config.GetAttribute(ctx, path.Root("search"), &search)
	resp.Diagnostics.Append(diag...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Call client API
	indices, sdkDiag := elasticsearch.GetIndices(ctx, &d.client, search)
	resp.Diagnostics.Append(utils.ConvertSDKDiagnosticsToFramework(sdkDiag)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Map response body to model
	for indexName, index := range indices {
		indexState := indexModel{
			Name:              types.StringValue(indexName),
			SortField:         types.SetValueMust(types.StringType, []attr.Value{}),
			SortOrder:         types.ListValueMust(types.StringType, []attr.Value{}),
			QueryDefaultField: types.SetValueMust(types.StringType, []attr.Value{}),
		}

		if uuid, ok := index.Settings["index.uuid"].(string); ok {
			indexState.ID = types.StringValue(uuid)
		} else {
			indexState.ID = types.StringValue(indexName)
		}

		// Map index settings
		if err := populateIndexFromSettings(ctx, utils.FlattenMap(index.Settings), &indexState); err != nil {
			resp.Diagnostics.AddError("unable to populate index from settings map", err.Error())
			return
		}
		// TODO: We ideally should set read settings to each field to detect changes
		// But for now, setting it will cause unexpected diff for the existing clients which use `settings`
		if index.Settings != nil {
			settings, err := json.Marshal(index.Settings)
			if err != nil {
				resp.Diagnostics.AddError("unable to marshal index settings", err.Error())
				return
			}
			indexState.SettingsRaw = types.StringValue(string(settings))
		}

		// Map index aliases
		for aliasName, alias := range index.Aliases {
			aliasState := aliasModel{
				Name:          types.StringValue(aliasName),
				IndexRouting:  types.StringValue(alias.IndexRouting),
				IsHidden:      types.BoolValue(alias.IsHidden),
				IsWriteIndex:  types.BoolValue(alias.IsWriteIndex),
				Routing:       types.StringValue(alias.Routing),
				SearchRouting: types.StringValue(alias.SearchRouting),
			}

			// Map index alias filter
			if alias.Filter != nil {
				filter, err := json.Marshal(alias.Filter)
				if err != nil {
					resp.Diagnostics.AddError("unable to marshal index alias filter", err.Error())
					return
				}
				aliasState.Filter = types.StringValue(string(filter))
			}

			indexState.Alias = append(indexState.Alias, aliasState)
		}

		// Map index mappings
		if index.Mappings != nil {
			mappings, err := json.Marshal(index.Mappings)
			if err != nil {
				resp.Diagnostics.AddError("unable to marshal index mappings", err.Error())
				return
			}
			indexState.Mappings = types.StringValue(string(mappings))
		}

		state.Indices = append(state.Indices, indexState)
	}

	state.ID = types.StringValue(search)

	// Set state
	diag = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diag...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func populateIndexFromSettings(ctx context.Context, settings map[string]interface{}, index interface{}) error {
	val := reflect.ValueOf(index).Elem()
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)
		tag := fieldType.Tag.Get("tfsdk")

		if tag == "" {
			continue
		}

		for key, value := range settings {
			if tag != utils.ConvertSettingsKeyToTFFieldKey(strings.TrimPrefix(key, "index.")) {
				continue
			}

			if val, ok := value.(string); ok {
				switch field.Type() {
				case reflect.TypeOf(types.String{}):
					field.Set(reflect.ValueOf(types.StringValue(val)))
				case reflect.TypeOf(types.Bool{}):
					if boolValue, err := strconv.ParseBool(val); err == nil {
						field.Set(reflect.ValueOf(types.BoolValue(boolValue)))
					} else {
						tflog.Warn(ctx, fmt.Sprintf("unable to parse boolean %v for %v", val, tag))
					}
				case reflect.TypeOf(types.Int32{}):
					if intValue, err := strconv.ParseInt(val, 10, 32); err == nil {
						field.Set(reflect.ValueOf(types.Int32Value(int32(intValue))))
					} else {
						tflog.Warn(ctx, fmt.Sprintf("unable to parse int32 %v for %v", val, tag))
					}
				case reflect.TypeOf(types.Int64{}):
					if intValue, err := strconv.ParseInt(val, 10, 64); err == nil {
						field.Set(reflect.ValueOf(types.Int64Value(intValue)))
					} else {
						tflog.Warn(ctx, fmt.Sprintf("unable to parse int64 %v for %v", val, tag))
					}
				case reflect.TypeOf(types.List{}):
					vals := strings.Split(val, ",")
					values := make([]attr.Value, len(vals))
					for i, v := range vals {
						values[i] = types.StringValue(v)
					}
					field.Set(reflect.ValueOf(types.ListValueMust(types.StringType, values)))
				case reflect.TypeOf(types.Set{}):
					vals := strings.Split(val, ",")
					values := make([]attr.Value, len(vals))
					for i, v := range vals {
						values[i] = types.StringValue(v)
					}
					field.Set(reflect.ValueOf(types.SetValueMust(types.StringType, values)))
				default:
					tflog.Warn(ctx, fmt.Sprintf(`Unsupport field type for "%s"`, tag))
				}
			}
		}
	}

	return nil
}
