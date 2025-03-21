/*
SLOs

OpenAPI schema for SLOs endpoints

API version: 1.1
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package slo

import (
	"encoding/json"
)

// checks if the IndicatorPropertiesTimesliceMetricParamsMetric type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &IndicatorPropertiesTimesliceMetricParamsMetric{}

// IndicatorPropertiesTimesliceMetricParamsMetric An object defining the metrics, equation, and threshold to determine if it's a good slice or not
type IndicatorPropertiesTimesliceMetricParamsMetric struct {
	// List of metrics with their name, aggregation type, and field.
	Metrics []IndicatorPropertiesTimesliceMetricParamsMetricMetricsInner `json:"metrics"`
	// The equation to calculate the metric.
	Equation string `json:"equation"`
	// The comparator to use to compare the equation to the threshold.
	Comparator string `json:"comparator"`
	// The threshold used to determine if the metric is a good slice or not.
	Threshold float64 `json:"threshold"`
}

// NewIndicatorPropertiesTimesliceMetricParamsMetric instantiates a new IndicatorPropertiesTimesliceMetricParamsMetric object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewIndicatorPropertiesTimesliceMetricParamsMetric(metrics []IndicatorPropertiesTimesliceMetricParamsMetricMetricsInner, equation string, comparator string, threshold float64) *IndicatorPropertiesTimesliceMetricParamsMetric {
	this := IndicatorPropertiesTimesliceMetricParamsMetric{}
	this.Metrics = metrics
	this.Equation = equation
	this.Comparator = comparator
	this.Threshold = threshold
	return &this
}

// NewIndicatorPropertiesTimesliceMetricParamsMetricWithDefaults instantiates a new IndicatorPropertiesTimesliceMetricParamsMetric object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewIndicatorPropertiesTimesliceMetricParamsMetricWithDefaults() *IndicatorPropertiesTimesliceMetricParamsMetric {
	this := IndicatorPropertiesTimesliceMetricParamsMetric{}
	return &this
}

// GetMetrics returns the Metrics field value
func (o *IndicatorPropertiesTimesliceMetricParamsMetric) GetMetrics() []IndicatorPropertiesTimesliceMetricParamsMetricMetricsInner {
	if o == nil {
		var ret []IndicatorPropertiesTimesliceMetricParamsMetricMetricsInner
		return ret
	}

	return o.Metrics
}

// GetMetricsOk returns a tuple with the Metrics field value
// and a boolean to check if the value has been set.
func (o *IndicatorPropertiesTimesliceMetricParamsMetric) GetMetricsOk() ([]IndicatorPropertiesTimesliceMetricParamsMetricMetricsInner, bool) {
	if o == nil {
		return nil, false
	}
	return o.Metrics, true
}

// SetMetrics sets field value
func (o *IndicatorPropertiesTimesliceMetricParamsMetric) SetMetrics(v []IndicatorPropertiesTimesliceMetricParamsMetricMetricsInner) {
	o.Metrics = v
}

// GetEquation returns the Equation field value
func (o *IndicatorPropertiesTimesliceMetricParamsMetric) GetEquation() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Equation
}

// GetEquationOk returns a tuple with the Equation field value
// and a boolean to check if the value has been set.
func (o *IndicatorPropertiesTimesliceMetricParamsMetric) GetEquationOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Equation, true
}

// SetEquation sets field value
func (o *IndicatorPropertiesTimesliceMetricParamsMetric) SetEquation(v string) {
	o.Equation = v
}

// GetComparator returns the Comparator field value
func (o *IndicatorPropertiesTimesliceMetricParamsMetric) GetComparator() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Comparator
}

// GetComparatorOk returns a tuple with the Comparator field value
// and a boolean to check if the value has been set.
func (o *IndicatorPropertiesTimesliceMetricParamsMetric) GetComparatorOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Comparator, true
}

// SetComparator sets field value
func (o *IndicatorPropertiesTimesliceMetricParamsMetric) SetComparator(v string) {
	o.Comparator = v
}

// GetThreshold returns the Threshold field value
func (o *IndicatorPropertiesTimesliceMetricParamsMetric) GetThreshold() float64 {
	if o == nil {
		var ret float64
		return ret
	}

	return o.Threshold
}

// GetThresholdOk returns a tuple with the Threshold field value
// and a boolean to check if the value has been set.
func (o *IndicatorPropertiesTimesliceMetricParamsMetric) GetThresholdOk() (*float64, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Threshold, true
}

// SetThreshold sets field value
func (o *IndicatorPropertiesTimesliceMetricParamsMetric) SetThreshold(v float64) {
	o.Threshold = v
}

func (o IndicatorPropertiesTimesliceMetricParamsMetric) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o IndicatorPropertiesTimesliceMetricParamsMetric) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["metrics"] = o.Metrics
	toSerialize["equation"] = o.Equation
	toSerialize["comparator"] = o.Comparator
	toSerialize["threshold"] = o.Threshold
	return toSerialize, nil
}

type NullableIndicatorPropertiesTimesliceMetricParamsMetric struct {
	value *IndicatorPropertiesTimesliceMetricParamsMetric
	isSet bool
}

func (v NullableIndicatorPropertiesTimesliceMetricParamsMetric) Get() *IndicatorPropertiesTimesliceMetricParamsMetric {
	return v.value
}

func (v *NullableIndicatorPropertiesTimesliceMetricParamsMetric) Set(val *IndicatorPropertiesTimesliceMetricParamsMetric) {
	v.value = val
	v.isSet = true
}

func (v NullableIndicatorPropertiesTimesliceMetricParamsMetric) IsSet() bool {
	return v.isSet
}

func (v *NullableIndicatorPropertiesTimesliceMetricParamsMetric) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableIndicatorPropertiesTimesliceMetricParamsMetric(val *IndicatorPropertiesTimesliceMetricParamsMetric) *NullableIndicatorPropertiesTimesliceMetricParamsMetric {
	return &NullableIndicatorPropertiesTimesliceMetricParamsMetric{value: val, isSet: true}
}

func (v NullableIndicatorPropertiesTimesliceMetricParamsMetric) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableIndicatorPropertiesTimesliceMetricParamsMetric) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
