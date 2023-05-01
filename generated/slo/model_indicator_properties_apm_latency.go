/*
SLOs

OpenAPI schema for SLOs endpoints

API version: 0.1
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package slo

import (
	"encoding/json"
)

// checks if the IndicatorPropertiesApmLatency type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &IndicatorPropertiesApmLatency{}

// IndicatorPropertiesApmLatency Defines properties for the APM latency indicator type
type IndicatorPropertiesApmLatency struct {
	Params IndicatorPropertiesApmLatencyParams `json:"params"`
	// The type of indicator.
	Type string `json:"type"`
}

// NewIndicatorPropertiesApmLatency instantiates a new IndicatorPropertiesApmLatency object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewIndicatorPropertiesApmLatency(params IndicatorPropertiesApmLatencyParams, type_ string) *IndicatorPropertiesApmLatency {
	this := IndicatorPropertiesApmLatency{}
	this.Params = params
	this.Type = type_
	return &this
}

// NewIndicatorPropertiesApmLatencyWithDefaults instantiates a new IndicatorPropertiesApmLatency object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewIndicatorPropertiesApmLatencyWithDefaults() *IndicatorPropertiesApmLatency {
	this := IndicatorPropertiesApmLatency{}
	return &this
}

// GetParams returns the Params field value
func (o *IndicatorPropertiesApmLatency) GetParams() IndicatorPropertiesApmLatencyParams {
	if o == nil {
		var ret IndicatorPropertiesApmLatencyParams
		return ret
	}

	return o.Params
}

// GetParamsOk returns a tuple with the Params field value
// and a boolean to check if the value has been set.
func (o *IndicatorPropertiesApmLatency) GetParamsOk() (*IndicatorPropertiesApmLatencyParams, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Params, true
}

// SetParams sets field value
func (o *IndicatorPropertiesApmLatency) SetParams(v IndicatorPropertiesApmLatencyParams) {
	o.Params = v
}

// GetType returns the Type field value
func (o *IndicatorPropertiesApmLatency) GetType() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Type
}

// GetTypeOk returns a tuple with the Type field value
// and a boolean to check if the value has been set.
func (o *IndicatorPropertiesApmLatency) GetTypeOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Type, true
}

// SetType sets field value
func (o *IndicatorPropertiesApmLatency) SetType(v string) {
	o.Type = v
}

func (o IndicatorPropertiesApmLatency) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o IndicatorPropertiesApmLatency) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["params"] = o.Params
	toSerialize["type"] = o.Type
	return toSerialize, nil
}

type NullableIndicatorPropertiesApmLatency struct {
	value *IndicatorPropertiesApmLatency
	isSet bool
}

func (v NullableIndicatorPropertiesApmLatency) Get() *IndicatorPropertiesApmLatency {
	return v.value
}

func (v *NullableIndicatorPropertiesApmLatency) Set(val *IndicatorPropertiesApmLatency) {
	v.value = val
	v.isSet = true
}

func (v NullableIndicatorPropertiesApmLatency) IsSet() bool {
	return v.isSet
}

func (v *NullableIndicatorPropertiesApmLatency) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableIndicatorPropertiesApmLatency(val *IndicatorPropertiesApmLatency) *NullableIndicatorPropertiesApmLatency {
	return &NullableIndicatorPropertiesApmLatency{value: val, isSet: true}
}

func (v NullableIndicatorPropertiesApmLatency) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableIndicatorPropertiesApmLatency) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
