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

package index

import indexparent "github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index"

// Elasticsearch index setting keys — thin aliases to the shared parent-package
// constants so existing code in this package continues to compile unchanged.
const (
	settingCodec                            = indexparent.SettingCodec
	settingLoadFixedBitsetFiltersEagerly    = indexparent.SettingLoadFixedBitsetFiltersEagerly
	settingMappingCoerce                    = indexparent.SettingMappingCoerce
	settingNumberOfReplicas                 = indexparent.SettingNumberOfReplicas
	settingAutoExpandReplicas               = indexparent.SettingAutoExpandReplicas
	settingMaxResultWindow                  = indexparent.SettingMaxResultWindow
	settingMaxInnerResultWindow             = indexparent.SettingMaxInnerResultWindow
	settingMaxRescoreWindow                 = indexparent.SettingMaxRescoreWindow
	settingMaxDocvalueFieldsSearch          = indexparent.SettingMaxDocvalueFieldsSearch
	settingMaxScriptFields                  = indexparent.SettingMaxScriptFields
	settingMaxNgramDiff                     = indexparent.SettingMaxNgramDiff
	settingMaxShingleDiff                   = indexparent.SettingMaxShingleDiff
	settingMaxRefreshListeners              = indexparent.SettingMaxRefreshListeners
	settingMaxTermsCount                    = indexparent.SettingMaxTermsCount
	settingMaxRegexLength                   = indexparent.SettingMaxRegexLength
	settingGCDeletes                        = indexparent.SettingGCDeletes
	settingDefaultPipeline                  = indexparent.SettingDefaultPipeline
	settingFinalPipeline                    = indexparent.SettingFinalPipeline
	settingNumberOfShards                   = indexparent.SettingNumberOfShards
	settingNumberOfRoutingShards            = indexparent.SettingNumberOfRoutingShards
	settingRoutingPartitionSize             = indexparent.SettingRoutingPartitionSize
	settingShardCheckOnStartup              = indexparent.SettingShardCheckOnStartup
	settingSortField                        = indexparent.SettingSortField
	settingSortOrder                        = indexparent.SettingSortOrder
	settingSortMissing                      = indexparent.SettingSortMissing
	settingSortMode                         = indexparent.SettingSortMode
	settingRefreshInterval                  = indexparent.SettingRefreshInterval
	settingSearchIdleAfter                  = indexparent.SettingSearchIdleAfter
	settingQueryDefaultField                = indexparent.SettingQueryDefaultField
	settingRoutingAllocationEnable          = indexparent.SettingRoutingAllocationEnable
	settingRoutingRebalanceEnable           = indexparent.SettingRoutingRebalanceEnable
	settingUnassignedNodeLeftDelayedTimeout = indexparent.SettingUnassignedNodeLeftDelayedTimeout
)

// Terraform schema attribute keys reused by the index resource and its sort
// support. These names match the Elasticsearch index sort and alias APIs.
const (
	attrName    = "name"
	attrFilter  = "filter"
	attrSetting = "setting"
	attrField   = "field"
	attrOrder   = "order"
	attrMissing = "missing"
	attrMode    = "mode"
)

// Sort order and tie-breaker tokens used in index sort plan modifiers and
// flattened state.
// importHydrationPrivateStateKey is set during ImportState and consumed on the
// following Read (hydrate all settings) and ModifyPlan (prune unconfigured fields).
const importHydrationPrivateStateKey = "import_hydration"

const (
	sortOrderAsc    = "asc"
	sortOrderDesc   = "desc"
	sortMissingLast = "_last"
	sortModeMax     = "max"
	sortModeMin     = "min"
)
