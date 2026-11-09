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

package rule

import (
	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	recoveryStrategyNoBreach  = "no_breach"
	recoveryStrategyCondition = "condition"
	recoveryStrategyQuery     = "query"
	recoveryStrategyManual    = "manual"

	noDataStrategyIgnore   = "ignore"
	noDataStrategyKeepLast = "keep_last"
	noDataStrategyResolve  = "resolve"
	noDataStrategyAlert    = "alert"
)

func (r *recoveryModel) toAPI() (*kbapi.KibanaHTTPAPIsAlertingRuleRecovery, diag.Diagnostics) {
	if r == nil {
		return nil, nil
	}

	var recovery kbapi.KibanaHTTPAPIsAlertingRuleRecovery
	var err error
	switch strategy := r.Strategy.ValueString(); strategy {
	case recoveryStrategyNoBreach:
		err = recovery.FromKibanaHTTPAPIsAlertingRuleRecoveryNoBreach(kbapi.KibanaHTTPAPIsAlertingRuleRecoveryNoBreach{})
	case recoveryStrategyCondition:
		err = recovery.FromKibanaHTTPAPIsAlertingRuleRecoveryCondition(kbapi.KibanaHTTPAPIsAlertingRuleRecoveryCondition{Segment: r.Segment.ValueString()})
	case recoveryStrategyQuery:
		err = recovery.FromKibanaHTTPAPIsAlertingRuleRecoveryQuery(kbapi.KibanaHTTPAPIsAlertingRuleRecoveryQuery{Query: r.Query.ValueString()})
	case recoveryStrategyManual:
		err = recovery.FromKibanaHTTPAPIsAlertingRuleRecoveryManual(kbapi.KibanaHTTPAPIsAlertingRuleRecoveryManual{})
	default:
		return nil, diag.Diagnostics{diag.NewErrorDiagnostic("Unsupported recovery strategy", "recovery.strategy "+strategy+" is not supported")}
	}
	if err != nil {
		return nil, diagutil.ErrDiag("Unable to build recovery", err)
	}
	return &recovery, nil
}

func (n *noDataModel) toAPI() (*kbapi.KibanaHTTPAPIsAlertingRuleNoData, diag.Diagnostics) {
	if n == nil {
		return nil, nil
	}

	var noData kbapi.KibanaHTTPAPIsAlertingRuleNoData
	var err error
	query := n.Query.ValueStringPointer()
	switch strategy := n.Strategy.ValueString(); strategy {
	case noDataStrategyIgnore:
		err = noData.FromKibanaHTTPAPIsAlertingRuleNoDataIgnore(kbapi.KibanaHTTPAPIsAlertingRuleNoDataIgnore{})
	case noDataStrategyKeepLast:
		err = noData.FromKibanaHTTPAPIsAlertingRuleNoDataKeepLast(kbapi.KibanaHTTPAPIsAlertingRuleNoDataKeepLast{Query: query})
	case noDataStrategyResolve:
		err = noData.FromKibanaHTTPAPIsAlertingRuleNoDataResolve(kbapi.KibanaHTTPAPIsAlertingRuleNoDataResolve{Query: query})
	default:
		return nil, diag.Diagnostics{diag.NewErrorDiagnostic("Unsupported no_data strategy", "no_data.strategy "+strategy+" is not supported")}
	}
	if err != nil {
		return nil, diagutil.ErrDiag("Unable to build no_data", err)
	}
	return &noData, nil
}

func recoveryFromAPI(recovery *kbapi.KibanaHTTPAPIsAlertingRuleRecovery) (*recoveryModel, diag.Diagnostics) {
	if recovery == nil {
		return nil, nil
	}

	strategy, err := recovery.Discriminator()
	if err != nil {
		return nil, diagutil.ErrDiag("Unable to read recovery strategy", err)
	}

	model := &recoveryModel{
		Strategy: types.StringValue(strategy),
		Segment:  types.StringNull(),
		Query:    types.StringNull(),
	}
	switch strategy {
	case recoveryStrategyNoBreach, recoveryStrategyManual:
	case recoveryStrategyCondition:
		condition, err := recovery.AsKibanaHTTPAPIsAlertingRuleRecoveryCondition()
		if err != nil {
			return nil, diagutil.ErrDiag("Unable to read recovery", err)
		}
		model.Segment = types.StringValue(condition.Segment)
	case recoveryStrategyQuery:
		query, err := recovery.AsKibanaHTTPAPIsAlertingRuleRecoveryQuery()
		if err != nil {
			return nil, diagutil.ErrDiag("Unable to read recovery", err)
		}
		model.Query = types.StringValue(query.Query)
	default:
		return nil, diag.Diagnostics{diag.NewErrorDiagnostic("Unsupported recovery strategy", "The API returned recovery.strategy "+strategy+", which this provider does not recognise.")}
	}
	return model, nil
}

func noDataFromAPI(noData *kbapi.KibanaHTTPAPIsAlertingRuleNoData) (*noDataModel, diag.Diagnostics) {
	if noData == nil {
		return nil, nil
	}

	strategy, err := noData.Discriminator()
	if err != nil {
		return nil, diagutil.ErrDiag("Unable to read no_data strategy", err)
	}

	model := &noDataModel{Strategy: types.StringValue(strategy), Query: types.StringNull()}
	switch strategy {
	case noDataStrategyIgnore:
	case noDataStrategyKeepLast:
		keepLast, err := noData.AsKibanaHTTPAPIsAlertingRuleNoDataKeepLast()
		if err != nil {
			return nil, diagutil.ErrDiag("Unable to read no_data", err)
		}
		model.Query = types.StringPointerValue(keepLast.Query)
	case noDataStrategyResolve:
		resolve, err := noData.AsKibanaHTTPAPIsAlertingRuleNoDataResolve()
		if err != nil {
			return nil, diagutil.ErrDiag("Unable to read no_data", err)
		}
		model.Query = types.StringPointerValue(resolve.Query)
	case noDataStrategyAlert:
		alert, err := noData.AsKibanaHTTPAPIsAlertingRuleNoDataAlert()
		if err != nil {
			return nil, diagutil.ErrDiag("Unable to read no_data", err)
		}
		model.Query = types.StringPointerValue(alert.Query)
	default:
		return nil, diag.Diagnostics{diag.NewErrorDiagnostic("Unsupported no_data strategy", "The API returned no_data.strategy "+strategy+", which this provider does not recognise.")}
	}
	return model, nil
}
