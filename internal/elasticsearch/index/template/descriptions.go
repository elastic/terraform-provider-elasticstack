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

package template

// Schema version for Plugin Framework state (REQ-041); bumped from SDK implicit 0.
const schemaVersion int64 = 1

// Resource and data source markdown descriptions (mirrors SDK embeds in internal/elasticsearch/index/descriptions).
const (
	mdDescIndexTemplateResource = "Creates or updates an index template. Index templates define settings, mappings, and aliases " +
		"that can be applied automatically to new indices. See, https://www.elastic.co/guide/en/elasticsearch/reference/current/indices-put-template.html"

	mdDescIndexTemplateDataSource = "Retrieves information about an existing index template definition. See, " +
		"https://www.elastic.co/guide/en/elasticsearch/reference/current/indices-get-template.html"
)

// Shared attribute and block descriptions — verbatim from Plugin SDK schema where noted.
const (
	descID          = "Internal identifier of the resource"
	descName        = "Name of the index template to create."
	descNameDataSrc = "The name of the index template."

	descComposedOf = "An ordered list of component template names."

	descIgnoreMissingComponentTemplates = "A list of component template names that are ignored if missing."

	descDataStreamBlock = "If this object is included, the template is used to create data streams and their backing indices. Supports an empty object."

	descDataStreamHidden = "If true, the data stream is hidden."

	descDataStreamAllowCustomRouting = "If `true`, the data stream supports custom routing. Defaults to `false`. Available only in **8.x**"

	descIndexPatterns = "Array of wildcard (*) expressions used to match the names of data streams and indices during creation."

	descMetadata = "Optional user metadata about the index template."

	descPriority = "Priority to determine index template precedence when a new data stream or index is created."

	descTemplateBlock = "Template to be applied. It may optionally include an aliases, mappings, lifecycle, or settings configuration."

	descAliasBlock = "Alias to add."

	descAliasName = "The alias name."

	descAliasFilter = "Query used to limit documents the alias can access."

	descAliasIndexRouting = "Value used to route indexing operations to a specific shard. If specified, this overwrites the `routing` value for indexing operations."

	descAliasIsHidden = "If true, the alias is hidden."

	descAliasIsWriteIndex = "If true, the index is the write index for the alias."

	descAliasRouting = "Value used to route indexing and search operations to a specific shard."

	descAliasSearchRouting = "Value used to route search operations to a specific shard. If specified, this overwrites the routing value for search operations."

	// Mirrors internal/elasticsearch/index/descriptions/index_template_mappings.md
	descTemplateMappings = "Mapping for fields in the index. Should be specified as a JSON object of field mappings. " +
		"See the documentation (https://www.elastic.co/guide/en/elasticsearch/reference/current/explicit-mapping.html) for more details"

	descTemplateSettings = "Configuration options for the index. See, " +
		"https://www.elastic.co/guide/en/elasticsearch/reference/current/index-modules.html#index-modules-settings"

	descLifecycleBlock = "Lifecycle of data stream. See, https://www.elastic.co/guide/en/elasticsearch/reference/current/data-stream-lifecycle.html"

	descLifecycleDataRetention = "The retention period of the data indexed in this data stream."

	descDataStreamOptionsBlock = "Options for data streams created by this template. Applied once at data stream creation time. Available only for Elasticsearch 9.1.0 and above."

	descDataStreamOptionsBlockDataSource = "Options for data streams created by this template. Available only for Elasticsearch 9.1.0 and above."

	descFailureStoreBlock = "Failure store configuration."

	descFailureStoreEnabled = "If true, document redirection to the failure store is enabled for new matching data streams."

	descFailureStoreLifecycleBlock = "Lifecycle configuration for the failure store."

	descFailureStoreDataRetention = "The retention period for failure store documents (e.g. \"30d\")."

	descVersion = "Version number used to manage index templates externally."
)

const (
	errSummaryMissingFailureStore = "Missing required failure_store block"
	errDetailMissingFailureStore  = "The `failure_store` block is required when `template.data_stream_options` is configured."
)
