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

// Terraform schema attribute keys shared across ingest processor data sources
// and the ingest pipeline resource.
const (
	attrJSON            = "json"
	attrField           = "field"
	attrValue           = "value"
	attrTargetField     = "target_field"
	attrIgnoreMissing   = "ignore_missing"
	attrAllowDuplicates = "allow_duplicates"
	attrSeparator       = "separator"
	attrOverride        = "override"
	attrProperties      = "properties"
	attrOnFailure       = "on_failure"
)

// Elasticsearch ingest processor type names returned by processor models.
const (
	processorTypeAppend = "append"
	processorTypeRemove = "remove"
	processorTypeSet    = "set"
)

// Schema descriptions reused across processor data sources and the pipeline resource.
const (
	descIdentifierWithPeriod = "Internal identifier of the resource."
	descJSONDataSource       = "JSON representation of this data source."
	descIgnoreMissingDocStop = "If `true` and `field` does not exist or is `null`, the processor quietly exits without modifying the document."
	descTargetFieldInPlace   = "The field to assign the converted value to, by default `field` is updated in-place."
)
