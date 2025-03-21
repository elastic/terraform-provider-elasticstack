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

// checks if the IndicatorPropertiesCustomMetric type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &IndicatorPropertiesCustomMetric{}

// IndicatorPropertiesCustomMetric Defines properties for a custom metric indicator type
type IndicatorPropertiesCustomMetric struct {
	Params IndicatorPropertiesCustomMetricParams `json:"params"`
	// The type of indicator.
	Type string `json:"type"`
}

// NewIndicatorPropertiesCustomMetric instantiates a new IndicatorPropertiesCustomMetric object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewIndicatorPropertiesCustomMetric(params IndicatorPropertiesCustomMetricParams, type_ string) *IndicatorPropertiesCustomMetric {
	this := IndicatorPropertiesCustomMetric{}
	this.Params = params
	this.Type = type_
	return &this
}

// NewIndicatorPropertiesCustomMetricWithDefaults instantiates a new IndicatorPropertiesCustomMetric object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewIndicatorPropertiesCustomMetricWithDefaults() *IndicatorPropertiesCustomMetric {
	this := IndicatorPropertiesCustomMetric{}
	return &this
}

// GetParams returns the Params field value
func (o *IndicatorPropertiesCustomMetric) GetParams() IndicatorPropertiesCustomMetricParams {
	if o == nil {
		var ret IndicatorPropertiesCustomMetricParams
		return ret
	}

	return o.Params
}

// GetParamsOk returns a tuple with the Params field value
// and a boolean to check if the value has been set.
func (o *IndicatorPropertiesCustomMetric) GetParamsOk() (*IndicatorPropertiesCustomMetricParams, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Params, true
}

// SetParams sets field value
func (o *IndicatorPropertiesCustomMetric) SetParams(v IndicatorPropertiesCustomMetricParams) {
	o.Params = v
}

// GetType returns the Type field value
func (o *IndicatorPropertiesCustomMetric) GetType() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Type
}

// GetTypeOk returns a tuple with the Type field value
// and a boolean to check if the value has been set.
func (o *IndicatorPropertiesCustomMetric) GetTypeOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Type, true
}

// SetType sets field value
func (o *IndicatorPropertiesCustomMetric) SetType(v string) {
	o.Type = v
}

func (o IndicatorPropertiesCustomMetric) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o IndicatorPropertiesCustomMetric) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["params"] = o.Params
	toSerialize["type"] = o.Type
	return toSerialize, nil
}

type NullableIndicatorPropertiesCustomMetric struct {
	value *IndicatorPropertiesCustomMetric
	isSet bool
}

func (v NullableIndicatorPropertiesCustomMetric) Get() *IndicatorPropertiesCustomMetric {
	return v.value
}

func (v *NullableIndicatorPropertiesCustomMetric) Set(val *IndicatorPropertiesCustomMetric) {
	v.value = val
	v.isSet = true
}

func (v NullableIndicatorPropertiesCustomMetric) IsSet() bool {
	return v.isSet
}

func (v *NullableIndicatorPropertiesCustomMetric) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableIndicatorPropertiesCustomMetric(val *IndicatorPropertiesCustomMetric) *NullableIndicatorPropertiesCustomMetric {
	return &NullableIndicatorPropertiesCustomMetric{value: val, isSet: true}
}

func (v NullableIndicatorPropertiesCustomMetric) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableIndicatorPropertiesCustomMetric) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
