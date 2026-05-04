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
type processorDropBody struct {
	CommonProcessorBody
}

// processorAppendBody is the JSON body for the append processor.
type processorAppendBody struct {
	CommonProcessorBody
	Field           string   `json:"field"`
	Value           []string `json:"value"`
	AllowDuplicates bool     `json:"allow_duplicates"`
	MediaType       string   `json:"media_type,omitempty"`
}

// processorScriptBody is the JSON body for the script processor.
type processorScriptBody struct {
	CommonProcessorBody
	Lang     string         `json:"lang,omitempty"`
	ScriptID string         `json:"id,omitempty"`
	Source   string         `json:"source,omitempty"`
	Params   map[string]any `json:"params,omitempty"`
}

// processorForeachBody is the JSON body for the foreach processor.
type processorForeachBody struct {
	CommonProcessorBody
	Field         string         `json:"field"`
	IgnoreMissing bool           `json:"ignore_missing"`
	Processor     map[string]any `json:"processor"`
}

// processorBytesBody is the JSON body for the bytes processor.
type processorBytesBody struct {
	CommonProcessorBody
	Field         string `json:"field"`
	TargetField   string `json:"target_field,omitempty"`
	IgnoreMissing bool   `json:"ignore_missing"`
}

// processorCircleBody is the JSON body for the circle processor.
type processorCircleBody struct {
	CommonProcessorBody
	Field         string  `json:"field"`
	TargetField   string  `json:"target_field,omitempty"`
	IgnoreMissing bool    `json:"ignore_missing"`
	ErrorDistance float64 `json:"error_distance"`
	ShapeType     string  `json:"shape_type"`
}

// processorCommunityIDBody is the JSON body for the community_id processor.
type processorCommunityIDBody struct {
	CommonProcessorBody
	SourceIP        string `json:"source_ip,omitempty"`
	SourcePort      *int   `json:"source_port,omitempty"`
	DestinationIP   string `json:"destination_ip,omitempty"`
	DestinationPort *int   `json:"destination_port,omitempty"`
	IanaNumber      *int   `json:"iana_number,omitempty"`
	IcmpType        *int   `json:"icmp_type,omitempty"`
	IcmpCode        *int   `json:"icmp_code,omitempty"`
	Transport       string `json:"transport,omitempty"`
	TargetField     string `json:"target_field,omitempty"`
	Seed            *int   `json:"seed,omitempty"`
	IgnoreMissing   bool   `json:"ignore_missing"`
}

// processorConvertBody is the JSON body for the convert processor.
type processorConvertBody struct {
	CommonProcessorBody
	Field         string `json:"field"`
	TargetField   string `json:"target_field,omitempty"`
	IgnoreMissing bool   `json:"ignore_missing"`
	Type          string `json:"type"`
}

// processorCSVBody is the JSON body for the csv processor.
type processorCSVBody struct {
	CommonProcessorBody
	Field         string   `json:"field"`
	TargetFields  []string `json:"target_fields"`
	IgnoreMissing bool     `json:"ignore_missing"`
	Separator     string   `json:"separator,omitempty"`
	Quote         string   `json:"quote,omitempty"`
	Trim          bool     `json:"trim"`
	EmptyValue    string   `json:"empty_value,omitempty"`
}

// processorDateBody is the JSON body for the date processor.
type processorDateBody struct {
	CommonProcessorBody
	Field        string   `json:"field"`
	TargetField  string   `json:"target_field,omitempty"`
	Formats      []string `json:"formats"`
	Timezone     string   `json:"timezone,omitempty"`
	Locale       string   `json:"locale,omitempty"`
	OutputFormat string   `json:"output_format,omitempty"`
}

// processorDateIndexNameBody is the JSON body for the date_index_name processor.
type processorDateIndexNameBody struct {
	CommonProcessorBody
	Field           string   `json:"field"`
	IndexNamePrefix string   `json:"index_name_prefix,omitempty"`
	DateRounding    string   `json:"date_rounding"`
	DateFormats     []string `json:"date_formats,omitempty"`
	Timezone        string   `json:"timezone,omitempty"`
	Locale          string   `json:"locale,omitempty"`
	IndexNameFormat string   `json:"index_name_format,omitempty"`
}

// processorDissectBody is the JSON body for the dissect processor.
type processorDissectBody struct {
	CommonProcessorBody
	Field           string `json:"field"`
	Pattern         string `json:"pattern"`
	AppendSeparator string `json:"append_separator"`
	IgnoreMissing   bool   `json:"ignore_missing"`
}

// processorDotExpanderBody is the JSON body for the dot_expander processor.
type processorDotExpanderBody struct {
	CommonProcessorBody
	Field    string `json:"field"`
	Path     string `json:"path,omitempty"`
	Override bool   `json:"override"`
}

// processorEnrichBody is the JSON body for the enrich processor.
type processorEnrichBody struct {
	CommonProcessorBody
	Field         string `json:"field"`
	TargetField   string `json:"target_field,omitempty"`
	IgnoreMissing bool   `json:"ignore_missing"`
	PolicyName    string `json:"policy_name"`
	Override      bool   `json:"override"`
	MaxMatches    int    `json:"max_matches"`
	ShapeRelation string `json:"shape_relation,omitempty"`
}

// processorFailBody is the JSON body for the fail processor.
type processorFailBody struct {
	CommonProcessorBody
	Message string `json:"message"`
}

// processorFingerprintBody is the JSON body for the fingerprint processor.
type processorFingerprintBody struct {
	CommonProcessorBody
	Fields        []string `json:"fields"`
	TargetField   string   `json:"target_field,omitempty"`
	IgnoreMissing bool     `json:"ignore_missing"`
	Salt          string   `json:"salt,omitempty"`
	Method        string   `json:"method,omitempty"`
}

// processorGeoIPBody is the JSON body for the geoip processor.
type processorGeoIPBody struct {
	CommonProcessorBody
	Field         string   `json:"field"`
	TargetField   string   `json:"target_field,omitempty"`
	IgnoreMissing bool     `json:"ignore_missing"`
	DatabaseFile  string   `json:"database_file,omitempty"`
	Properties    []string `json:"properties,omitempty"`
	FirstOnly     bool     `json:"first_only"`
}

// processorGrokBody is the JSON body for the grok processor.
type processorGrokBody struct {
	CommonProcessorBody
	Field              string            `json:"field"`
	Patterns           []string          `json:"patterns"`
	PatternDefinitions map[string]string `json:"pattern_definitions,omitempty"`
	EcsCompatibility   string            `json:"ecs_compatibility,omitempty"`
	TraceMatch         bool              `json:"trace_match"`
	IgnoreMissing      bool              `json:"ignore_missing"`
}

// processorGsubBody is the JSON body for the gsub processor.
type processorGsubBody struct {
	CommonProcessorBody
	Field         string `json:"field"`
	TargetField   string `json:"target_field,omitempty"`
	IgnoreMissing bool   `json:"ignore_missing"`
	Pattern       string `json:"pattern"`
	Replacement   string `json:"replacement"`
}

// processorHTMLStripBody is the JSON body for the html_strip processor.
type processorHTMLStripBody struct {
	CommonProcessorBody
	Field         string `json:"field"`
	TargetField   string `json:"target_field,omitempty"`
	IgnoreMissing bool   `json:"ignore_missing"`
}

// processorInferenceInputOutputBody is the JSON body for the inference input_output.
type processorInferenceInputOutputBody struct {
	InputField  string `json:"input_field"`
	OutputField string `json:"output_field,omitempty"`
}

// processorInferenceBody is the JSON body for the inference processor.
type processorInferenceBody struct {
	CommonProcessorBody
	ModelID     string                             `json:"model_id"`
	InputOutput *processorInferenceInputOutputBody `json:"input_output,omitempty"`
	FieldMap    map[string]string                  `json:"field_map,omitempty"`
	TargetField string                             `json:"target_field,omitempty"`
}

// processorJoinBody is the JSON body for the join processor.
type processorJoinBody struct {
	CommonProcessorBody
	Field       string `json:"field"`
	Separator   string `json:"separator"`
	TargetField string `json:"target_field,omitempty"`
}

// processorJSONBody is the JSON body for the json processor.
type processorJSONBody struct {
	CommonProcessorBody
	Field                     string `json:"field"`
	TargetField               string `json:"target_field,omitempty"`
	AddToRoot                 *bool  `json:"add_to_root,omitempty"`
	AddToRootConflictStrategy string `json:"add_to_root_conflict_strategy,omitempty"`
	AllowDuplicateKeys        *bool  `json:"allow_duplicate_keys,omitempty"`
}

// processorKVBody is the JSON body for the kv processor.
type processorKVBody struct {
	CommonProcessorBody
	Field         string   `json:"field"`
	TargetField   string   `json:"target_field,omitempty"`
	IgnoreMissing bool     `json:"ignore_missing"`
	FieldSplit    string   `json:"field_split"`
	ValueSplit    string   `json:"value_split"`
	IncludeKeys   []string `json:"include_keys,omitempty"`
	ExcludeKeys   []string `json:"exclude_keys,omitempty"`
	Prefix        string   `json:"prefix,omitempty"`
	TrimKey       string   `json:"trim_key,omitempty"`
	TrimValue     string   `json:"trim_value,omitempty"`
	StripBrackets bool     `json:"strip_brackets"`
}

// processorLowercaseBody is the JSON body for the lowercase processor.
type processorLowercaseBody struct {
	CommonProcessorBody
	Field         string `json:"field"`
	TargetField   string `json:"target_field,omitempty"`
	IgnoreMissing bool   `json:"ignore_missing"`
}

// processorNetworkDirectionBody is the JSON body for the network_direction processor.
type processorNetworkDirectionBody struct {
	CommonProcessorBody
	SourceIP              string   `json:"source_ip,omitempty"`
	DestinationIP         string   `json:"destination_ip,omitempty"`
	TargetField           string   `json:"target_field,omitempty"`
	InternalNetworks      []string `json:"internal_networks,omitempty"`
	InternalNetworksField string   `json:"internal_networks_field,omitempty"`
	IgnoreMissing         bool     `json:"ignore_missing"`
}

// processorPipelineBody is the JSON body for the pipeline processor.
type processorPipelineBody struct {
	CommonProcessorBody
	Name string `json:"name"`
}

// processorRegisteredDomainBody is the JSON body for the registered_domain processor.
type processorRegisteredDomainBody struct {
	CommonProcessorBody
	Field         string `json:"field"`
	TargetField   string `json:"target_field,omitempty"`
	IgnoreMissing bool   `json:"ignore_missing"`
}

// processorRemoveBody is the JSON body for the remove processor.
type processorRemoveBody struct {
	CommonProcessorBody
	Field         []string `json:"field"`
	IgnoreMissing bool     `json:"ignore_missing"`
}

// processorRenameBody is the JSON body for the rename processor.
type processorRenameBody struct {
	CommonProcessorBody
	Field         string `json:"field"`
	TargetField   string `json:"target_field"`
	IgnoreMissing bool   `json:"ignore_missing"`
}

// processorRerouteBody is the JSON body for the reroute processor.
type processorRerouteBody struct {
	CommonProcessorBody
	Destination string `json:"destination,omitempty"`
	Dataset     string `json:"dataset,omitempty"`
	Namespace   string `json:"namespace,omitempty"`
}

// processorSetBody is the JSON body for the set processor.
type processorSetBody struct {
	CommonProcessorBody
	Field            string `json:"field"`
	Value            string `json:"value,omitempty"`
	CopyFrom         string `json:"copy_from,omitempty"`
	Override         bool   `json:"override"`
	IgnoreEmptyValue bool   `json:"ignore_empty_value"`
	MediaType        string `json:"media_type,omitempty"`
}

// processorSetSecurityUserBody is the JSON body for the set_security_user processor.
type processorSetSecurityUserBody struct {
	CommonProcessorBody
	Field      string   `json:"field"`
	Properties []string `json:"properties,omitempty"`
}

// processorSortBody is the JSON body for the sort processor.
type processorSortBody struct {
	CommonProcessorBody
	Field       string `json:"field"`
	Order       string `json:"order,omitempty"`
	TargetField string `json:"target_field,omitempty"`
}

// processorSplitBody is the JSON body for the split processor.
type processorSplitBody struct {
	CommonProcessorBody
	Field            string `json:"field"`
	TargetField      string `json:"target_field,omitempty"`
	IgnoreMissing    bool   `json:"ignore_missing"`
	Separator        string `json:"separator"`
	PreserveTrailing bool   `json:"preserve_trailing"`
}

// processorTrimBody is the JSON body for the trim processor.
type processorTrimBody struct {
	CommonProcessorBody
	Field         string `json:"field"`
	TargetField   string `json:"target_field,omitempty"`
	IgnoreMissing bool   `json:"ignore_missing"`
}

// processorUppercaseBody is the JSON body for the uppercase processor.
type processorUppercaseBody struct {
	CommonProcessorBody
	Field         string `json:"field"`
	TargetField   string `json:"target_field,omitempty"`
	IgnoreMissing bool   `json:"ignore_missing"`
}

// processorURIPartsBody is the JSON body for the uri_parts processor.
type processorURIPartsBody struct {
	CommonProcessorBody
	Field              string `json:"field"`
	TargetField        string `json:"target_field,omitempty"`
	KeepOriginal       bool   `json:"keep_original"`
	RemoveIfSuccessful bool   `json:"remove_if_successful"`
}

// processorURLDecodeBody is the JSON body for the urldecode processor.
type processorURLDecodeBody struct {
	CommonProcessorBody
	Field         string `json:"field"`
	TargetField   string `json:"target_field,omitempty"`
	IgnoreMissing bool   `json:"ignore_missing"`
}

// processorUserAgentBody is the JSON body for the user_agent processor.
type processorUserAgentBody struct {
	CommonProcessorBody
	Field             string   `json:"field"`
	TargetField       string   `json:"target_field,omitempty"`
	IgnoreMissing     bool     `json:"ignore_missing"`
	RegexFile         string   `json:"regex_file,omitempty"`
	Properties        []string `json:"properties,omitempty"`
	ExtractDeviceType *bool    `json:"extract_device_type,omitempty"`
}
