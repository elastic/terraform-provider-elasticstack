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

package watch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Data is the Terraform state/plan model for the watch resource.
type Data struct {
	ID                      types.String         `tfsdk:"id"`
	ElasticsearchConnection types.List           `tfsdk:"elasticsearch_connection"`
	WatchID                 types.String         `tfsdk:"watch_id"`
	Active                  types.Bool           `tfsdk:"active"`
	Trigger                 jsontypes.Normalized `tfsdk:"trigger"`
	Input                   jsontypes.Normalized `tfsdk:"input"`
	Condition               jsontypes.Normalized `tfsdk:"condition"`
	Actions                 jsontypes.Normalized `tfsdk:"actions"`
	Metadata                jsontypes.Normalized `tfsdk:"metadata"`
	Transform               jsontypes.Normalized `tfsdk:"transform"`
	ThrottlePeriodInMillis  types.Int64          `tfsdk:"throttle_period_in_millis"`
}

// toPutModel converts the Terraform state into a models.PutWatch for the API.
func (d *Data) toPutModel(_ context.Context) (*models.PutWatch, diag.Diagnostics) {
	var diags diag.Diagnostics

	put := &models.PutWatch{
		WatchID: d.WatchID.ValueString(),
		Active:  d.Active.ValueBool(),
	}

	if err := json.Unmarshal([]byte(d.Trigger.ValueString()), &put.Body.Trigger); err != nil {
		diags.AddError("Invalid trigger JSON", fmt.Sprintf("Error parsing trigger: %s", err))
		return nil, diags
	}

	if err := json.Unmarshal([]byte(d.Input.ValueString()), &put.Body.Input); err != nil {
		diags.AddError("Invalid input JSON", fmt.Sprintf("Error parsing input: %s", err))
		return nil, diags
	}

	if err := json.Unmarshal([]byte(d.Condition.ValueString()), &put.Body.Condition); err != nil {
		diags.AddError("Invalid condition JSON", fmt.Sprintf("Error parsing condition: %s", err))
		return nil, diags
	}

	if err := json.Unmarshal([]byte(d.Actions.ValueString()), &put.Body.Actions); err != nil {
		diags.AddError("Invalid actions JSON", fmt.Sprintf("Error parsing actions: %s", err))
		return nil, diags
	}

	if err := json.Unmarshal([]byte(d.Metadata.ValueString()), &put.Body.Metadata); err != nil {
		diags.AddError("Invalid metadata JSON", fmt.Sprintf("Error parsing metadata: %s", err))
		return nil, diags
	}

	if !d.Transform.IsNull() && !d.Transform.IsUnknown() {
		var transform map[string]any
		if err := json.Unmarshal([]byte(d.Transform.ValueString()), &transform); err != nil {
			diags.AddError("Invalid transform JSON", fmt.Sprintf("Error parsing transform: %s", err))
			return nil, diags
		}
		put.Body.Transform = transform
	}

	put.Body.ThrottlePeriodInMillis = int(d.ThrottlePeriodInMillis.ValueInt64())

	return put, diags
}

// marshalCompact marshals v to compact JSON. It returns an error if marshaling
// or compaction fails. Using json.Compact ensures the stored value is always
// compact regardless of what format the API returned (e.g. pretty-printed on
// older Elasticsearch versions).
func marshalCompact(v any) (string, error) {
	raw, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := json.Compact(&buf, raw); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// fromAPIModel populates the Data model from an API watch response.
func (d *Data) fromAPIModel(_ context.Context, watch *models.Watch) diag.Diagnostics {
	var diags diag.Diagnostics

	d.WatchID = types.StringValue(watch.WatchID)
	d.Active = types.BoolValue(watch.Status.State.Active)

	if watch.Body.Trigger == nil {
		diags.AddError("API Response Error", "Watch trigger is missing from API response")
		return diags
	}
	trigger, err := marshalCompact(watch.Body.Trigger)
	if err != nil {
		diags.AddError("JSON Marshal Error", fmt.Sprintf("Error marshaling trigger: %s", err))
		return diags
	}
	d.Trigger = jsontypes.NewNormalizedValue(trigger)

	if watch.Body.Input == nil {
		d.Input = jsontypes.NewNormalizedValue(`{"none":{}}`)
	} else {
		input, err := marshalCompact(watch.Body.Input)
		if err != nil {
			diags.AddError("JSON Marshal Error", fmt.Sprintf("Error marshaling input: %s", err))
			return diags
		}
		d.Input = jsontypes.NewNormalizedValue(input)
	}

	if watch.Body.Condition == nil {
		d.Condition = jsontypes.NewNormalizedValue(`{"always":{}}`)
	} else {
		condition, err := marshalCompact(watch.Body.Condition)
		if err != nil {
			diags.AddError("JSON Marshal Error", fmt.Sprintf("Error marshaling condition: %s", err))
			return diags
		}
		d.Condition = jsontypes.NewNormalizedValue(condition)
	}

	if watch.Body.Actions == nil {
		d.Actions = jsontypes.NewNormalizedValue(`{}`)
	} else {
		actions, err := marshalCompact(watch.Body.Actions)
		if err != nil {
			diags.AddError("JSON Marshal Error", fmt.Sprintf("Error marshaling actions: %s", err))
			return diags
		}
		d.Actions = jsontypes.NewNormalizedValue(actions)
	}

	if watch.Body.Metadata == nil {
		d.Metadata = jsontypes.NewNormalizedValue(`{}`)
	} else {
		metadata, err := marshalCompact(watch.Body.Metadata)
		if err != nil {
			diags.AddError("JSON Marshal Error", fmt.Sprintf("Error marshaling metadata: %s", err))
			return diags
		}
		d.Metadata = jsontypes.NewNormalizedValue(metadata)
	}

	if watch.Body.Transform != nil {
		transform, err := marshalCompact(watch.Body.Transform)
		if err != nil {
			diags.AddError("JSON Marshal Error", fmt.Sprintf("Error marshaling transform: %s", err))
			return diags
		}
		d.Transform = jsontypes.NewNormalizedValue(transform)
	} else {
		d.Transform = jsontypes.NewNormalizedNull()
	}

	d.ThrottlePeriodInMillis = types.Int64Value(int64(watch.Body.ThrottlePeriodInMillis))

	return diags
}
