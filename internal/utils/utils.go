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

package schemautil

import (
	"context"
	"fmt"
	maps0 "maps"
	"strings"

	providerSchema "github.com/elastic/terraform-provider-elasticstack/internal/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func MergeSchemaMaps(maps ...map[string]*schema.Schema) map[string]*schema.Schema {
	result := make(map[string]*schema.Schema)
	for _, m := range maps {
		maps0.Copy(result, m)
	}
	return result
}

const connectionKeyName = "elasticsearch_connection"

// AddConnectionSchema returns the common connection schema for all Elasticsearch resources,
// which defines the fields used to configure API access.
func AddConnectionSchema(providedSchema map[string]*schema.Schema) {
	providedSchema[connectionKeyName] = providerSchema.GetEsConnectionSchema(connectionKeyName, false)
}

func ExpandIndividuallyDefinedSettings(ctx context.Context, d *schema.ResourceData, settingsKeys map[string]schema.ValueType) map[string]any {
	settings := make(map[string]any)
	for key := range settingsKeys {
		tfFieldKey := ConvertSettingsKeyToTFFieldKey(key)
		if raw, ok := d.GetOk(tfFieldKey); ok {
			switch field := raw.(type) {
			case *schema.Set:
				settings[key] = field.List()
			default:
				settings[key] = raw
			}
			tflog.Trace(ctx, fmt.Sprintf("expandIndividuallyDefinedSettings: settingsKey:%+v tfFieldKey:%+v value:%+v, %+v", key, tfFieldKey, raw, settings))
		}
	}
	return settings
}

func ConvertSettingsKeyToTFFieldKey(settingKey string) string {
	return strings.ReplaceAll(settingKey, ".", "_")
}
