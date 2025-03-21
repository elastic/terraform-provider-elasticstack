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

// checks if the IndicatorPropertiesCustomMetricParamsGoodMetricsInnerOneOf type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &IndicatorPropertiesCustomMetricParamsGoodMetricsInnerOneOf{}

// IndicatorPropertiesCustomMetricParamsGoodMetricsInnerOneOf struct for IndicatorPropertiesCustomMetricParamsGoodMetricsInnerOneOf
type IndicatorPropertiesCustomMetricParamsGoodMetricsInnerOneOf struct {
	// The name of the metric. Only valid options are A-Z
	Name string `json:"name"`
	// The aggregation type of the metric.
	Aggregation string `json:"aggregation"`
	// The field of the metric.
	Field string `json:"field"`
	// The filter to apply to the metric.
	Filter *string `json:"filter,omitempty"`
}

// NewIndicatorPropertiesCustomMetricParamsGoodMetricsInnerOneOf instantiates a new IndicatorPropertiesCustomMetricParamsGoodMetricsInnerOneOf object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewIndicatorPropertiesCustomMetricParamsGoodMetricsInnerOneOf(name string, aggregation string, field string) *IndicatorPropertiesCustomMetricParamsGoodMetricsInnerOneOf {
	this := IndicatorPropertiesCustomMetricParamsGoodMetricsInnerOneOf{}
	this.Name = name
	this.Aggregation = aggregation
	this.Field = field
	return &this
}

// NewIndicatorPropertiesCustomMetricParamsGoodMetricsInnerOneOfWithDefaults instantiates a new IndicatorPropertiesCustomMetricParamsGoodMetricsInnerOneOf object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewIndicatorPropertiesCustomMetricParamsGoodMetricsInnerOneOfWithDefaults() *IndicatorPropertiesCustomMetricParamsGoodMetricsInnerOneOf {
	this := IndicatorPropertiesCustomMetricParamsGoodMetricsInnerOneOf{}
	return &this
}

// GetName returns the Name field value
func (o *IndicatorPropertiesCustomMetricParamsGoodMetricsInnerOneOf) GetName() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Name
}

// GetNameOk returns a tuple with the Name field value
// and a boolean to check if the value has been set.
func (o *IndicatorPropertiesCustomMetricParamsGoodMetricsInnerOneOf) GetNameOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Name, true
}

// SetName sets field value
func (o *IndicatorPropertiesCustomMetricParamsGoodMetricsInnerOneOf) SetName(v string) {
	o.Name = v
}

// GetAggregation returns the Aggregation field value
func (o *IndicatorPropertiesCustomMetricParamsGoodMetricsInnerOneOf) GetAggregation() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Aggregation
}

// GetAggregationOk returns a tuple with the Aggregation field value
// and a boolean to check if the value has been set.
func (o *IndicatorPropertiesCustomMetricParamsGoodMetricsInnerOneOf) GetAggregationOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Aggregation, true
}

// SetAggregation sets field value
func (o *IndicatorPropertiesCustomMetricParamsGoodMetricsInnerOneOf) SetAggregation(v string) {
	o.Aggregation = v
}

// GetField returns the Field field value
func (o *IndicatorPropertiesCustomMetricParamsGoodMetricsInnerOneOf) GetField() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Field
}

// GetFieldOk returns a tuple with the Field field value
// and a boolean to check if the value has been set.
func (o *IndicatorPropertiesCustomMetricParamsGoodMetricsInnerOneOf) GetFieldOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Field, true
}

// SetField sets field value
func (o *IndicatorPropertiesCustomMetricParamsGoodMetricsInnerOneOf) SetField(v string) {
	o.Field = v
}

// GetFilter returns the Filter field value if set, zero value otherwise.
func (o *IndicatorPropertiesCustomMetricParamsGoodMetricsInnerOneOf) GetFilter() string {
	if o == nil || IsNil(o.Filter) {
		var ret string
		return ret
	}
	return *o.Filter
}

// GetFilterOk returns a tuple with the Filter field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *IndicatorPropertiesCustomMetricParamsGoodMetricsInnerOneOf) GetFilterOk() (*string, bool) {
	if o == nil || IsNil(o.Filter) {
		return nil, false
	}
	return o.Filter, true
}

// HasFilter returns a boolean if a field has been set.
func (o *IndicatorPropertiesCustomMetricParamsGoodMetricsInnerOneOf) HasFilter() bool {
	if o != nil && !IsNil(o.Filter) {
		return true
	}

	return false
}

// SetFilter gets a reference to the given string and assigns it to the Filter field.
func (o *IndicatorPropertiesCustomMetricParamsGoodMetricsInnerOneOf) SetFilter(v string) {
	o.Filter = &v
}

func (o IndicatorPropertiesCustomMetricParamsGoodMetricsInnerOneOf) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o IndicatorPropertiesCustomMetricParamsGoodMetricsInnerOneOf) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["name"] = o.Name
	toSerialize["aggregation"] = o.Aggregation
	toSerialize["field"] = o.Field
	if !IsNil(o.Filter) {
		toSerialize["filter"] = o.Filter
	}
	return toSerialize, nil
}

type NullableIndicatorPropertiesCustomMetricParamsGoodMetricsInnerOneOf struct {
	value *IndicatorPropertiesCustomMetricParamsGoodMetricsInnerOneOf
	isSet bool
}

func (v NullableIndicatorPropertiesCustomMetricParamsGoodMetricsInnerOneOf) Get() *IndicatorPropertiesCustomMetricParamsGoodMetricsInnerOneOf {
	return v.value
}

func (v *NullableIndicatorPropertiesCustomMetricParamsGoodMetricsInnerOneOf) Set(val *IndicatorPropertiesCustomMetricParamsGoodMetricsInnerOneOf) {
	v.value = val
	v.isSet = true
}

func (v NullableIndicatorPropertiesCustomMetricParamsGoodMetricsInnerOneOf) IsSet() bool {
	return v.isSet
}

func (v *NullableIndicatorPropertiesCustomMetricParamsGoodMetricsInnerOneOf) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableIndicatorPropertiesCustomMetricParamsGoodMetricsInnerOneOf(val *IndicatorPropertiesCustomMetricParamsGoodMetricsInnerOneOf) *NullableIndicatorPropertiesCustomMetricParamsGoodMetricsInnerOneOf {
	return &NullableIndicatorPropertiesCustomMetricParamsGoodMetricsInnerOneOf{value: val, isSet: true}
}

func (v NullableIndicatorPropertiesCustomMetricParamsGoodMetricsInnerOneOf) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableIndicatorPropertiesCustomMetricParamsGoodMetricsInnerOneOf) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
