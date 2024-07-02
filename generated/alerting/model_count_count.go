/*
Alerting

OpenAPI schema for alerting endpoints

API version: 0.2
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package alerting

import (
	"encoding/json"
)

// checks if the CountCount type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &CountCount{}

// CountCount struct for CountCount
type CountCount struct {
	Comparator *string  `json:"comparator,omitempty"`
	Value      *float32 `json:"value,omitempty"`
}

// NewCountCount instantiates a new CountCount object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewCountCount() *CountCount {
	this := CountCount{}
	return &this
}

// NewCountCountWithDefaults instantiates a new CountCount object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewCountCountWithDefaults() *CountCount {
	this := CountCount{}
	return &this
}

// GetComparator returns the Comparator field value if set, zero value otherwise.
func (o *CountCount) GetComparator() string {
	if o == nil || IsNil(o.Comparator) {
		var ret string
		return ret
	}
	return *o.Comparator
}

// GetComparatorOk returns a tuple with the Comparator field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *CountCount) GetComparatorOk() (*string, bool) {
	if o == nil || IsNil(o.Comparator) {
		return nil, false
	}
	return o.Comparator, true
}

// HasComparator returns a boolean if a field has been set.
func (o *CountCount) HasComparator() bool {
	if o != nil && !IsNil(o.Comparator) {
		return true
	}

	return false
}

// SetComparator gets a reference to the given string and assigns it to the Comparator field.
func (o *CountCount) SetComparator(v string) {
	o.Comparator = &v
}

// GetValue returns the Value field value if set, zero value otherwise.
func (o *CountCount) GetValue() float32 {
	if o == nil || IsNil(o.Value) {
		var ret float32
		return ret
	}
	return *o.Value
}

// GetValueOk returns a tuple with the Value field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *CountCount) GetValueOk() (*float32, bool) {
	if o == nil || IsNil(o.Value) {
		return nil, false
	}
	return o.Value, true
}

// HasValue returns a boolean if a field has been set.
func (o *CountCount) HasValue() bool {
	if o != nil && !IsNil(o.Value) {
		return true
	}

	return false
}

// SetValue gets a reference to the given float32 and assigns it to the Value field.
func (o *CountCount) SetValue(v float32) {
	o.Value = &v
}

func (o CountCount) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o CountCount) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	if !IsNil(o.Comparator) {
		toSerialize["comparator"] = o.Comparator
	}
	if !IsNil(o.Value) {
		toSerialize["value"] = o.Value
	}
	return toSerialize, nil
}

type NullableCountCount struct {
	value *CountCount
	isSet bool
}

func (v NullableCountCount) Get() *CountCount {
	return v.value
}

func (v *NullableCountCount) Set(val *CountCount) {
	v.value = val
	v.isSet = true
}

func (v NullableCountCount) IsSet() bool {
	return v.isSet
}

func (v *NullableCountCount) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableCountCount(val *CountCount) *NullableCountCount {
	return &NullableCountCount{value: val, isSet: true}
}

func (v NullableCountCount) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableCountCount) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
