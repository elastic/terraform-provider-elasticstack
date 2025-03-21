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

// checks if the IndicatorPropertiesCustomKqlParams type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &IndicatorPropertiesCustomKqlParams{}

// IndicatorPropertiesCustomKqlParams An object containing the indicator parameters.
type IndicatorPropertiesCustomKqlParams struct {
	// The index or index pattern to use
	Index string `json:"index"`
	// The kibana data view id to use, primarily used to include data view runtime mappings. Make sure to save SLO again if you add/update run time fields to the data view and if those fields are being used in slo queries.
	DataViewId *string             `json:"dataViewId,omitempty"`
	Filter     *KqlWithFilters     `json:"filter,omitempty"`
	Good       KqlWithFiltersGood  `json:"good"`
	Total      KqlWithFiltersTotal `json:"total"`
	// The timestamp field used in the source indice.
	TimestampField string `json:"timestampField"`
}

// NewIndicatorPropertiesCustomKqlParams instantiates a new IndicatorPropertiesCustomKqlParams object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewIndicatorPropertiesCustomKqlParams(index string, good KqlWithFiltersGood, total KqlWithFiltersTotal, timestampField string) *IndicatorPropertiesCustomKqlParams {
	this := IndicatorPropertiesCustomKqlParams{}
	this.Index = index
	this.Good = good
	this.Total = total
	this.TimestampField = timestampField
	return &this
}

// NewIndicatorPropertiesCustomKqlParamsWithDefaults instantiates a new IndicatorPropertiesCustomKqlParams object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewIndicatorPropertiesCustomKqlParamsWithDefaults() *IndicatorPropertiesCustomKqlParams {
	this := IndicatorPropertiesCustomKqlParams{}
	return &this
}

// GetIndex returns the Index field value
func (o *IndicatorPropertiesCustomKqlParams) GetIndex() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Index
}

// GetIndexOk returns a tuple with the Index field value
// and a boolean to check if the value has been set.
func (o *IndicatorPropertiesCustomKqlParams) GetIndexOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Index, true
}

// SetIndex sets field value
func (o *IndicatorPropertiesCustomKqlParams) SetIndex(v string) {
	o.Index = v
}

// GetDataViewId returns the DataViewId field value if set, zero value otherwise.
func (o *IndicatorPropertiesCustomKqlParams) GetDataViewId() string {
	if o == nil || IsNil(o.DataViewId) {
		var ret string
		return ret
	}
	return *o.DataViewId
}

// GetDataViewIdOk returns a tuple with the DataViewId field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *IndicatorPropertiesCustomKqlParams) GetDataViewIdOk() (*string, bool) {
	if o == nil || IsNil(o.DataViewId) {
		return nil, false
	}
	return o.DataViewId, true
}

// HasDataViewId returns a boolean if a field has been set.
func (o *IndicatorPropertiesCustomKqlParams) HasDataViewId() bool {
	if o != nil && !IsNil(o.DataViewId) {
		return true
	}

	return false
}

// SetDataViewId gets a reference to the given string and assigns it to the DataViewId field.
func (o *IndicatorPropertiesCustomKqlParams) SetDataViewId(v string) {
	o.DataViewId = &v
}

// GetFilter returns the Filter field value if set, zero value otherwise.
func (o *IndicatorPropertiesCustomKqlParams) GetFilter() KqlWithFilters {
	if o == nil || IsNil(o.Filter) {
		var ret KqlWithFilters
		return ret
	}
	return *o.Filter
}

// GetFilterOk returns a tuple with the Filter field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *IndicatorPropertiesCustomKqlParams) GetFilterOk() (*KqlWithFilters, bool) {
	if o == nil || IsNil(o.Filter) {
		return nil, false
	}
	return o.Filter, true
}

// HasFilter returns a boolean if a field has been set.
func (o *IndicatorPropertiesCustomKqlParams) HasFilter() bool {
	if o != nil && !IsNil(o.Filter) {
		return true
	}

	return false
}

// SetFilter gets a reference to the given KqlWithFilters and assigns it to the Filter field.
func (o *IndicatorPropertiesCustomKqlParams) SetFilter(v KqlWithFilters) {
	o.Filter = &v
}

// GetGood returns the Good field value
func (o *IndicatorPropertiesCustomKqlParams) GetGood() KqlWithFiltersGood {
	if o == nil {
		var ret KqlWithFiltersGood
		return ret
	}

	return o.Good
}

// GetGoodOk returns a tuple with the Good field value
// and a boolean to check if the value has been set.
func (o *IndicatorPropertiesCustomKqlParams) GetGoodOk() (*KqlWithFiltersGood, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Good, true
}

// SetGood sets field value
func (o *IndicatorPropertiesCustomKqlParams) SetGood(v KqlWithFiltersGood) {
	o.Good = v
}

// GetTotal returns the Total field value
func (o *IndicatorPropertiesCustomKqlParams) GetTotal() KqlWithFiltersTotal {
	if o == nil {
		var ret KqlWithFiltersTotal
		return ret
	}

	return o.Total
}

// GetTotalOk returns a tuple with the Total field value
// and a boolean to check if the value has been set.
func (o *IndicatorPropertiesCustomKqlParams) GetTotalOk() (*KqlWithFiltersTotal, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Total, true
}

// SetTotal sets field value
func (o *IndicatorPropertiesCustomKqlParams) SetTotal(v KqlWithFiltersTotal) {
	o.Total = v
}

// GetTimestampField returns the TimestampField field value
func (o *IndicatorPropertiesCustomKqlParams) GetTimestampField() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.TimestampField
}

// GetTimestampFieldOk returns a tuple with the TimestampField field value
// and a boolean to check if the value has been set.
func (o *IndicatorPropertiesCustomKqlParams) GetTimestampFieldOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.TimestampField, true
}

// SetTimestampField sets field value
func (o *IndicatorPropertiesCustomKqlParams) SetTimestampField(v string) {
	o.TimestampField = v
}

func (o IndicatorPropertiesCustomKqlParams) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o IndicatorPropertiesCustomKqlParams) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["index"] = o.Index
	if !IsNil(o.DataViewId) {
		toSerialize["dataViewId"] = o.DataViewId
	}
	if !IsNil(o.Filter) {
		toSerialize["filter"] = o.Filter
	}
	toSerialize["good"] = o.Good
	toSerialize["total"] = o.Total
	toSerialize["timestampField"] = o.TimestampField
	return toSerialize, nil
}

type NullableIndicatorPropertiesCustomKqlParams struct {
	value *IndicatorPropertiesCustomKqlParams
	isSet bool
}

func (v NullableIndicatorPropertiesCustomKqlParams) Get() *IndicatorPropertiesCustomKqlParams {
	return v.value
}

func (v *NullableIndicatorPropertiesCustomKqlParams) Set(val *IndicatorPropertiesCustomKqlParams) {
	v.value = val
	v.isSet = true
}

func (v NullableIndicatorPropertiesCustomKqlParams) IsSet() bool {
	return v.isSet
}

func (v *NullableIndicatorPropertiesCustomKqlParams) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableIndicatorPropertiesCustomKqlParams(val *IndicatorPropertiesCustomKqlParams) *NullableIndicatorPropertiesCustomKqlParams {
	return &NullableIndicatorPropertiesCustomKqlParams{value: val, isSet: true}
}

func (v NullableIndicatorPropertiesCustomKqlParams) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableIndicatorPropertiesCustomKqlParams) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
