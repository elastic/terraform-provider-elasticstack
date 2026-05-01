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

package ingest

import (
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// CommonProcessorBody holds the JSON-tagged fields shared by every ingest
// processor. Embed this struct in per-processor body shapes.
type CommonProcessorBody struct {
	Description   string           `json:"description,omitempty"`
	If            string           `json:"if,omitempty"`
	IgnoreFailure bool             `json:"ignore_failure"`
	OnFailure     []map[string]any `json:"on_failure,omitempty"`
	Tag           string           `json:"tag,omitempty"`
}

// toCommonProcessorBody translates a CommonProcessorModel into a
// CommonProcessorBody. It returns any diagnostics collected while parsing
// on_failure JSON values.
//

func toCommonProcessorBody(model CommonProcessorModel) (CommonProcessorBody, diag.Diagnostics) {
	var body CommonProcessorBody
	var diags diag.Diagnostics

	if IsKnown(model.Description) {
		body.Description = model.Description.ValueString()
	}
	if IsKnown(model.If) {
		body.If = model.If.ValueString()
	}
	if IsKnown(model.IgnoreFailure) {
		body.IgnoreFailure = model.IgnoreFailure.ValueBool()
	}
	if IsKnown(model.OnFailure) {
		for _, elem := range model.OnFailure.Elements() {
			norm, ok := elem.(jsontypes.Normalized)
			if !ok {
				diags.AddError("Invalid on_failure element type", "expected jsontypes.Normalized")
				continue
			}
			if !IsKnown(norm) {
				diags.AddError("Unknown on_failure element", "on_failure elements cannot be unknown")
				continue
			}
			var item map[string]any
			if err := json.Unmarshal([]byte(norm.ValueString()), &item); err != nil {
				diags.AddError("Failed to parse on_failure JSON", err.Error())
				continue
			}
			body.OnFailure = append(body.OnFailure, item)
		}
	}
	if IsKnown(model.Tag) {
		body.Tag = model.Tag.ValueString()
	}

	return body, diags
}

// processorDropBody is the JSON body for the drop processor.
//

type processorDropBody struct {
	CommonProcessorBody
}

// processorAppendBody is the JSON body for the append processor.
//

type processorAppendBody struct {
	CommonProcessorBody
	Field           string   `json:"field"`
	Value           []string `json:"value"`
	AllowDuplicates bool     `json:"allow_duplicates"`
	MediaType       string   `json:"media_type,omitempty"`
}

// processorScriptBody is the JSON body for the script processor.
//

type processorScriptBody struct {
	CommonProcessorBody
	Lang     string         `json:"lang,omitempty"`
	ScriptID string         `json:"id,omitempty"`
	Source   string         `json:"source,omitempty"`
	Params   map[string]any `json:"params,omitempty"`
}

// processorForeachBody is the JSON body for the foreach processor.
//

type processorForeachBody struct {
	CommonProcessorBody
	Field         string         `json:"field"`
	IgnoreMissing bool           `json:"ignore_missing"`
	Processor     map[string]any `json:"processor"`
}
