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

// checks if the IndicatorPropertiesHistogramParams type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &IndicatorPropertiesHistogramParams{}

// IndicatorPropertiesHistogramParams An object containing the indicator parameters.
type IndicatorPropertiesHistogramParams struct {
	// The index or index pattern to use
	Index string `json:"index"`
	// The kibana data view id to use, primarily used to include data view runtime mappings. Make sure to save SLO again if you add/update run time fields to the data view and if those fields are being used in slo queries.
	DataViewId *string `json:"dataViewId,omitempty"`
	// the KQL query to filter the documents with.
	Filter *string `json:"filter,omitempty"`
	// The timestamp field used in the source indice.
	TimestampField string                                  `json:"timestampField"`
	Good           IndicatorPropertiesHistogramParamsGood  `json:"good"`
	Total          IndicatorPropertiesHistogramParamsTotal `json:"total"`
}

// NewIndicatorPropertiesHistogramParams instantiates a new IndicatorPropertiesHistogramParams object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewIndicatorPropertiesHistogramParams(index string, timestampField string, good IndicatorPropertiesHistogramParamsGood, total IndicatorPropertiesHistogramParamsTotal) *IndicatorPropertiesHistogramParams {
	this := IndicatorPropertiesHistogramParams{}
	this.Index = index
	this.TimestampField = timestampField
	this.Good = good
	this.Total = total
	return &this
}

// NewIndicatorPropertiesHistogramParamsWithDefaults instantiates a new IndicatorPropertiesHistogramParams object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewIndicatorPropertiesHistogramParamsWithDefaults() *IndicatorPropertiesHistogramParams {
	this := IndicatorPropertiesHistogramParams{}
	return &this
}

// GetIndex returns the Index field value
func (o *IndicatorPropertiesHistogramParams) GetIndex() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Index
}

// GetIndexOk returns a tuple with the Index field value
// and a boolean to check if the value has been set.
func (o *IndicatorPropertiesHistogramParams) GetIndexOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Index, true
}

// SetIndex sets field value
func (o *IndicatorPropertiesHistogramParams) SetIndex(v string) {
	o.Index = v
}

// GetDataViewId returns the DataViewId field value if set, zero value otherwise.
func (o *IndicatorPropertiesHistogramParams) GetDataViewId() string {
	if o == nil || IsNil(o.DataViewId) {
		var ret string
		return ret
	}
	return *o.DataViewId
}

// GetDataViewIdOk returns a tuple with the DataViewId field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *IndicatorPropertiesHistogramParams) GetDataViewIdOk() (*string, bool) {
	if o == nil || IsNil(o.DataViewId) {
		return nil, false
	}
	return o.DataViewId, true
}

// HasDataViewId returns a boolean if a field has been set.
func (o *IndicatorPropertiesHistogramParams) HasDataViewId() bool {
	if o != nil && !IsNil(o.DataViewId) {
		return true
	}

	return false
}

// SetDataViewId gets a reference to the given string and assigns it to the DataViewId field.
func (o *IndicatorPropertiesHistogramParams) SetDataViewId(v string) {
	o.DataViewId = &v
}

// GetFilter returns the Filter field value if set, zero value otherwise.
func (o *IndicatorPropertiesHistogramParams) GetFilter() string {
	if o == nil || IsNil(o.Filter) {
		var ret string
		return ret
	}
	return *o.Filter
}

// GetFilterOk returns a tuple with the Filter field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *IndicatorPropertiesHistogramParams) GetFilterOk() (*string, bool) {
	if o == nil || IsNil(o.Filter) {
		return nil, false
	}
	return o.Filter, true
}

// HasFilter returns a boolean if a field has been set.
func (o *IndicatorPropertiesHistogramParams) HasFilter() bool {
	if o != nil && !IsNil(o.Filter) {
		return true
	}

	return false
}

// SetFilter gets a reference to the given string and assigns it to the Filter field.
func (o *IndicatorPropertiesHistogramParams) SetFilter(v string) {
	o.Filter = &v
}

// GetTimestampField returns the TimestampField field value
func (o *IndicatorPropertiesHistogramParams) GetTimestampField() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.TimestampField
}

// GetTimestampFieldOk returns a tuple with the TimestampField field value
// and a boolean to check if the value has been set.
func (o *IndicatorPropertiesHistogramParams) GetTimestampFieldOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.TimestampField, true
}

// SetTimestampField sets field value
func (o *IndicatorPropertiesHistogramParams) SetTimestampField(v string) {
	o.TimestampField = v
}

// GetGood returns the Good field value
func (o *IndicatorPropertiesHistogramParams) GetGood() IndicatorPropertiesHistogramParamsGood {
	if o == nil {
		var ret IndicatorPropertiesHistogramParamsGood
		return ret
	}

	return o.Good
}

// GetGoodOk returns a tuple with the Good field value
// and a boolean to check if the value has been set.
func (o *IndicatorPropertiesHistogramParams) GetGoodOk() (*IndicatorPropertiesHistogramParamsGood, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Good, true
}

// SetGood sets field value
func (o *IndicatorPropertiesHistogramParams) SetGood(v IndicatorPropertiesHistogramParamsGood) {
	o.Good = v
}

// GetTotal returns the Total field value
func (o *IndicatorPropertiesHistogramParams) GetTotal() IndicatorPropertiesHistogramParamsTotal {
	if o == nil {
		var ret IndicatorPropertiesHistogramParamsTotal
		return ret
	}

	return o.Total
}

// GetTotalOk returns a tuple with the Total field value
// and a boolean to check if the value has been set.
func (o *IndicatorPropertiesHistogramParams) GetTotalOk() (*IndicatorPropertiesHistogramParamsTotal, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Total, true
}

// SetTotal sets field value
func (o *IndicatorPropertiesHistogramParams) SetTotal(v IndicatorPropertiesHistogramParamsTotal) {
	o.Total = v
}

func (o IndicatorPropertiesHistogramParams) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o IndicatorPropertiesHistogramParams) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["index"] = o.Index
	if !IsNil(o.DataViewId) {
		toSerialize["dataViewId"] = o.DataViewId
	}
	if !IsNil(o.Filter) {
		toSerialize["filter"] = o.Filter
	}
	toSerialize["timestampField"] = o.TimestampField
	toSerialize["good"] = o.Good
	toSerialize["total"] = o.Total
	return toSerialize, nil
}

type NullableIndicatorPropertiesHistogramParams struct {
	value *IndicatorPropertiesHistogramParams
	isSet bool
}

func (v NullableIndicatorPropertiesHistogramParams) Get() *IndicatorPropertiesHistogramParams {
	return v.value
}

func (v *NullableIndicatorPropertiesHistogramParams) Set(val *IndicatorPropertiesHistogramParams) {
	v.value = val
	v.isSet = true
}

func (v NullableIndicatorPropertiesHistogramParams) IsSet() bool {
	return v.isSet
}

func (v *NullableIndicatorPropertiesHistogramParams) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableIndicatorPropertiesHistogramParams(val *IndicatorPropertiesHistogramParams) *NullableIndicatorPropertiesHistogramParams {
	return &NullableIndicatorPropertiesHistogramParams{value: val, isSet: true}
}

func (v NullableIndicatorPropertiesHistogramParams) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableIndicatorPropertiesHistogramParams) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
