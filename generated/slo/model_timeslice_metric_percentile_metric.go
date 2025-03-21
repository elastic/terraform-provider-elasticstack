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

// checks if the TimesliceMetricPercentileMetric type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &TimesliceMetricPercentileMetric{}

// TimesliceMetricPercentileMetric struct for TimesliceMetricPercentileMetric
type TimesliceMetricPercentileMetric struct {
	// The name of the metric. Only valid options are A-Z
	Name string `json:"name"`
	// The aggregation type of the metric. Only valid option is \"percentile\"
	Aggregation string `json:"aggregation"`
	// The field of the metric.
	Field string `json:"field"`
	// The percentile value.
	Percentile float64 `json:"percentile"`
	// The filter to apply to the metric.
	Filter *string `json:"filter,omitempty"`
}

// NewTimesliceMetricPercentileMetric instantiates a new TimesliceMetricPercentileMetric object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewTimesliceMetricPercentileMetric(name string, aggregation string, field string, percentile float64) *TimesliceMetricPercentileMetric {
	this := TimesliceMetricPercentileMetric{}
	this.Name = name
	this.Aggregation = aggregation
	this.Field = field
	this.Percentile = percentile
	return &this
}

// NewTimesliceMetricPercentileMetricWithDefaults instantiates a new TimesliceMetricPercentileMetric object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewTimesliceMetricPercentileMetricWithDefaults() *TimesliceMetricPercentileMetric {
	this := TimesliceMetricPercentileMetric{}
	return &this
}

// GetName returns the Name field value
func (o *TimesliceMetricPercentileMetric) GetName() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Name
}

// GetNameOk returns a tuple with the Name field value
// and a boolean to check if the value has been set.
func (o *TimesliceMetricPercentileMetric) GetNameOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Name, true
}

// SetName sets field value
func (o *TimesliceMetricPercentileMetric) SetName(v string) {
	o.Name = v
}

// GetAggregation returns the Aggregation field value
func (o *TimesliceMetricPercentileMetric) GetAggregation() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Aggregation
}

// GetAggregationOk returns a tuple with the Aggregation field value
// and a boolean to check if the value has been set.
func (o *TimesliceMetricPercentileMetric) GetAggregationOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Aggregation, true
}

// SetAggregation sets field value
func (o *TimesliceMetricPercentileMetric) SetAggregation(v string) {
	o.Aggregation = v
}

// GetField returns the Field field value
func (o *TimesliceMetricPercentileMetric) GetField() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Field
}

// GetFieldOk returns a tuple with the Field field value
// and a boolean to check if the value has been set.
func (o *TimesliceMetricPercentileMetric) GetFieldOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Field, true
}

// SetField sets field value
func (o *TimesliceMetricPercentileMetric) SetField(v string) {
	o.Field = v
}

// GetPercentile returns the Percentile field value
func (o *TimesliceMetricPercentileMetric) GetPercentile() float64 {
	if o == nil {
		var ret float64
		return ret
	}

	return o.Percentile
}

// GetPercentileOk returns a tuple with the Percentile field value
// and a boolean to check if the value has been set.
func (o *TimesliceMetricPercentileMetric) GetPercentileOk() (*float64, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Percentile, true
}

// SetPercentile sets field value
func (o *TimesliceMetricPercentileMetric) SetPercentile(v float64) {
	o.Percentile = v
}

// GetFilter returns the Filter field value if set, zero value otherwise.
func (o *TimesliceMetricPercentileMetric) GetFilter() string {
	if o == nil || IsNil(o.Filter) {
		var ret string
		return ret
	}
	return *o.Filter
}

// GetFilterOk returns a tuple with the Filter field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *TimesliceMetricPercentileMetric) GetFilterOk() (*string, bool) {
	if o == nil || IsNil(o.Filter) {
		return nil, false
	}
	return o.Filter, true
}

// HasFilter returns a boolean if a field has been set.
func (o *TimesliceMetricPercentileMetric) HasFilter() bool {
	if o != nil && !IsNil(o.Filter) {
		return true
	}

	return false
}

// SetFilter gets a reference to the given string and assigns it to the Filter field.
func (o *TimesliceMetricPercentileMetric) SetFilter(v string) {
	o.Filter = &v
}

func (o TimesliceMetricPercentileMetric) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o TimesliceMetricPercentileMetric) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["name"] = o.Name
	toSerialize["aggregation"] = o.Aggregation
	toSerialize["field"] = o.Field
	toSerialize["percentile"] = o.Percentile
	if !IsNil(o.Filter) {
		toSerialize["filter"] = o.Filter
	}
	return toSerialize, nil
}

type NullableTimesliceMetricPercentileMetric struct {
	value *TimesliceMetricPercentileMetric
	isSet bool
}

func (v NullableTimesliceMetricPercentileMetric) Get() *TimesliceMetricPercentileMetric {
	return v.value
}

func (v *NullableTimesliceMetricPercentileMetric) Set(val *TimesliceMetricPercentileMetric) {
	v.value = val
	v.isSet = true
}

func (v NullableTimesliceMetricPercentileMetric) IsSet() bool {
	return v.isSet
}

func (v *NullableTimesliceMetricPercentileMetric) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableTimesliceMetricPercentileMetric(val *TimesliceMetricPercentileMetric) *NullableTimesliceMetricPercentileMetric {
	return &NullableTimesliceMetricPercentileMetric{value: val, isSet: true}
}

func (v NullableTimesliceMetricPercentileMetric) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableTimesliceMetricPercentileMetric) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
