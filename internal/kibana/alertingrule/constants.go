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

package alertingrule

// Kibana alerting rule type IDs used in params validation and compatibility
// defaults.
const (
	ruleTypeIndexThreshold          = ".index-threshold"
	ruleTypeESQuery                 = ".es-query"
	ruleTypeApmTransactionErrorRate = "apm.transaction_error_rate"
	ruleTypeUptimeMonitorStatus     = "xpack.uptime.alerts.monitorStatus"
)

// Terraform schema attribute keys.
const (
	attrTags          = "tags"
	attrParams        = "params"
	attrNotifyWhen    = "notify_when"
	attrRuleTypeID    = "rule_type_id"
	attrEnabled       = "enabled"
	attrThrottle      = "throttle"
	blockFrequency    = "frequency"
	blockAlertsFilter = "alerts_filter"
)

// JSON params keys used across rule types.
const (
	paramsKeyGroupBy = "groupBy"
	paramsKeyAggType = "aggType"
	paramsKeySize    = "size"
)

// Default params values applied for backward-compatible API behavior.
const (
	paramsGroupByAll   = "all"
	paramsAggTypeCount = "count"
)
