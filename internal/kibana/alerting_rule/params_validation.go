package alerting_rule

import (
	"fmt"
)

type paramsValidatorFunc func(map[string]interface{}) []string

var ruleTypeParamsValidators = map[string]paramsValidatorFunc{
	"apm.rules.anomaly":                 validateApmAnomalyParams,
	"apm.error_rate":                    validateApmErrorCountParams,
	"apm.transaction_duration":          validateApmTransactionDurationParams,
	"apm.transaction_error_rate":        validateApmTransactionErrorRateParams,
	".es-query":                         validateEsQueryParams,
	".index-threshold":                  validateIndexThresholdParams,
	"metrics.alert.inventory.threshold": validateInfraInventoryParams,
	"logs.alert.document.count":         validateLogThresholdParams,
	"metrics.alert.threshold":           validateInfraMetricThresholdParams,
	"slo.rules.burnRate":                validateSloBurnRateParams,
	"xpack.uptime.alerts.tls":           validateSyntheticsUptimeTLSParams,
	"xpack.uptime.alerts.monitorStatus": validateSyntheticsMonitorStatusParams,
}

func validateRuleParams(ruleTypeID string, params map[string]interface{}) []string {
	validator, ok := ruleTypeParamsValidators[ruleTypeID]
	if !ok {
		// Backward compatible fallback for custom/unknown rules.
		return nil
	}

	return validator(params)
}

func validateApmAnomalyParams(params map[string]interface{}) []string {
	var errs []string
	errs = append(errs, requireNumber(params, "windowSize")...)
	errs = append(errs, requireStringEnum(params, "windowUnit", "m", "h", "d")...)
	errs = append(errs, requireString(params, "environment")...)
	errs = append(errs, requireStringEnum(params, "anomalySeverityType", "critical", "major", "minor", "warning")...)
	return errs
}

func validateApmErrorCountParams(params map[string]interface{}) []string {
	var errs []string
	errs = append(errs, requireNumber(params, "windowSize")...)
	errs = append(errs, requireStringEnum(params, "windowUnit", "m", "h", "d")...)
	errs = append(errs, requireNumber(params, "threshold")...)
	errs = append(errs, requireString(params, "environment")...)
	return errs
}

func validateApmTransactionDurationParams(params map[string]interface{}) []string {
	var errs []string
	errs = append(errs, requireNumber(params, "windowSize")...)
	errs = append(errs, requireStringEnum(params, "windowUnit", "m", "h", "d")...)
	errs = append(errs, requireNumber(params, "threshold")...)
	errs = append(errs, requireString(params, "environment")...)
	errs = append(errs, requireStringEnum(params, "aggregationType", "avg", "95th", "99th")...)
	return errs
}

func validateApmTransactionErrorRateParams(params map[string]interface{}) []string {
	var errs []string
	errs = append(errs, requireNumber(params, "windowSize")...)
	errs = append(errs, requireStringEnum(params, "windowUnit", "m", "h", "d")...)
	errs = append(errs, requireNumber(params, "threshold")...)
	errs = append(errs, requireString(params, "environment")...)
	return errs
}

func validateIndexThresholdParams(params map[string]interface{}) []string {
	var errs []string
	errs = append(errs, requireStringSlice(params, "index")...)
	errs = append(errs, requireNumberSlice(params, "threshold")...)
	errs = append(errs, requireStringEnum(params, "thresholdComparator", ">", ">=", "<", "<=", "between", "notBetween")...)
	errs = append(errs, requireString(params, "timeField")...)
	errs = append(errs, requireNumber(params, "timeWindowSize")...)
	errs = append(errs, requireStringEnum(params, "timeWindowUnit", "s", "m", "h", "d")...)
	return errs
}

func validateEsQueryParams(params map[string]interface{}) []string {
	candidates := []paramsValidatorFunc{
		validateEsQueryDSLParams,
		validateEsQueryESQLParams,
		validateEsQueryKQLParams,
	}

	var best []string
	for _, candidate := range candidates {
		errs := candidate(params)
		if len(errs) == 0 {
			return nil
		}
		if best == nil || len(errs) < len(best) {
			best = errs
		}
	}

	return best
}

func validateEsQueryDSLParams(params map[string]interface{}) []string {
	var errs []string
	errs = append(errs, requireString(params, "esQuery")...)
	errs = append(errs, requireStringOrStringSlice(params, "index")...)
	errs = append(errs, requireNumberSlice(params, "threshold")...)
	errs = append(errs, requireStringEnum(params, "thresholdComparator", ">", ">=", "<", "<=", "between", "notBetween")...)
	errs = append(errs, requireString(params, "timeField")...)
	errs = append(errs, requireNumber(params, "timeWindowSize")...)
	errs = append(errs, requireStringEnum(params, "timeWindowUnit", "s", "m", "h", "d")...)
	return errs
}

func validateEsQueryESQLParams(params map[string]interface{}) []string {
	var errs []string
	errs = append(errs, requireNestedString(params, "esqlQuery", "esql")...)
	errs = append(errs, requireStringEnum(params, "searchType", "esqlQuery")...)
	errs = append(errs, requireNumber(params, "size")...)
	errs = append(errs, requireNumberSlice(params, "threshold")...)
	errs = append(errs, requireStringEnum(params, "thresholdComparator", ">")...)
	errs = append(errs, requireNumber(params, "timeWindowSize")...)
	errs = append(errs, requireStringEnum(params, "timeWindowUnit", "s", "m", "h", "d")...)
	return errs
}

func validateEsQueryKQLParams(params map[string]interface{}) []string {
	var errs []string
	errs = append(errs, requireStringEnum(params, "searchType", "searchSource")...)
	errs = append(errs, requireNumber(params, "size")...)
	errs = append(errs, requireNumberSlice(params, "threshold")...)
	errs = append(errs, requireStringEnum(params, "thresholdComparator", ">", ">=", "<", "<=", "between", "notBetween")...)
	errs = append(errs, requireNumber(params, "timeWindowSize")...)
	errs = append(errs, requireStringEnum(params, "timeWindowUnit", "s", "m", "h", "d")...)
	return errs
}

func validateInfraInventoryParams(map[string]interface{}) []string {
	return nil
}

func validateLogThresholdParams(params map[string]interface{}) []string {
	var errs []string
	errs = append(errs, requireObject(params, "count")...)
	errs = append(errs, requireNumber(params, "timeSize")...)
	errs = append(errs, requireStringEnum(params, "timeUnit", "s", "m", "h", "d")...)
	errs = append(errs, requireObject(params, "logView")...)
	return errs
}

func validateInfraMetricThresholdParams(map[string]interface{}) []string {
	return nil
}

func validateSloBurnRateParams(map[string]interface{}) []string {
	return nil
}

func validateSyntheticsUptimeTLSParams(map[string]interface{}) []string {
	return nil
}

func validateSyntheticsMonitorStatusParams(params map[string]interface{}) []string {
	var errs []string
	errs = append(errs, requireNumber(params, "numTimes")...)
	errs = append(errs, requireBool(params, "shouldCheckStatus")...)
	errs = append(errs, requireBool(params, "shouldCheckAvailability")...)
	return errs
}

func requireString(params map[string]interface{}, key string) []string {
	value, ok := params[key]
	if !ok {
		return []string{fmt.Sprintf("params.%s is required", key)}
	}
	if _, ok := value.(string); !ok {
		return []string{fmt.Sprintf("params.%s must be a string", key)}
	}
	return nil
}

func requireStringEnum(params map[string]interface{}, key string, allowed ...string) []string {
	value, ok := params[key]
	if !ok {
		return []string{fmt.Sprintf("params.%s is required", key)}
	}

	asString, ok := value.(string)
	if !ok {
		return []string{fmt.Sprintf("params.%s must be a string", key)}
	}

	for _, v := range allowed {
		if asString == v {
			return nil
		}
	}

	return []string{fmt.Sprintf("params.%s must be one of %v", key, allowed)}
}

func requireNumber(params map[string]interface{}, key string) []string {
	value, ok := params[key]
	if !ok {
		return []string{fmt.Sprintf("params.%s is required", key)}
	}
	if !isNumber(value) {
		return []string{fmt.Sprintf("params.%s must be a number", key)}
	}
	return nil
}

func requireBool(params map[string]interface{}, key string) []string {
	value, ok := params[key]
	if !ok {
		return []string{fmt.Sprintf("params.%s is required", key)}
	}
	if _, ok := value.(bool); !ok {
		return []string{fmt.Sprintf("params.%s must be a boolean", key)}
	}
	return nil
}

func requireObject(params map[string]interface{}, key string) []string {
	value, ok := params[key]
	if !ok {
		return []string{fmt.Sprintf("params.%s is required", key)}
	}
	if _, ok := value.(map[string]interface{}); !ok {
		return []string{fmt.Sprintf("params.%s must be an object", key)}
	}
	return nil
}

func requireStringSlice(params map[string]interface{}, key string) []string {
	value, ok := params[key]
	if !ok {
		return []string{fmt.Sprintf("params.%s is required", key)}
	}
	items, ok := value.([]interface{})
	if !ok {
		return []string{fmt.Sprintf("params.%s must be an array of strings", key)}
	}
	for _, item := range items {
		if _, ok := item.(string); !ok {
			return []string{fmt.Sprintf("params.%s must be an array of strings", key)}
		}
	}
	return nil
}

func requireNumberSlice(params map[string]interface{}, key string) []string {
	value, ok := params[key]
	if !ok {
		return []string{fmt.Sprintf("params.%s is required", key)}
	}
	items, ok := value.([]interface{})
	if !ok {
		return []string{fmt.Sprintf("params.%s must be an array of numbers", key)}
	}
	for _, item := range items {
		if !isNumber(item) {
			return []string{fmt.Sprintf("params.%s must be an array of numbers", key)}
		}
	}
	return nil
}

func requireStringOrStringSlice(params map[string]interface{}, key string) []string {
	value, ok := params[key]
	if !ok {
		return []string{fmt.Sprintf("params.%s is required", key)}
	}
	if _, ok := value.(string); ok {
		return nil
	}

	items, ok := value.([]interface{})
	if !ok {
		return []string{fmt.Sprintf("params.%s must be a string or an array of strings", key)}
	}
	for _, item := range items {
		if _, ok := item.(string); !ok {
			return []string{fmt.Sprintf("params.%s must be a string or an array of strings", key)}
		}
	}
	return nil
}

func requireNestedString(params map[string]interface{}, objectKey string, childKey string) []string {
	value, ok := params[objectKey]
	if !ok {
		return []string{fmt.Sprintf("params.%s is required", objectKey)}
	}

	objectValue, ok := value.(map[string]interface{})
	if !ok {
		return []string{fmt.Sprintf("params.%s must be an object", objectKey)}
	}

	childValue, ok := objectValue[childKey]
	if !ok {
		return []string{fmt.Sprintf("params.%s.%s is required", objectKey, childKey)}
	}

	if _, ok := childValue.(string); !ok {
		return []string{fmt.Sprintf("params.%s.%s must be a string", objectKey, childKey)}
	}

	return nil
}

func isNumber(value interface{}) bool {
	switch value.(type) {
	case int, int8, int16, int32, int64:
		return true
	case uint, uint8, uint16, uint32, uint64:
		return true
	case float32, float64:
		return true
	default:
		return false
	}
}
