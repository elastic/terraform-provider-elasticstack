/*
Data views

OpenAPI schema for data view endpoints

API version: 0.1
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package data_views

import (
	"encoding/json"
)

// checks if the GetDefaultDataView200Response type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &GetDefaultDataView200Response{}

// GetDefaultDataView200Response struct for GetDefaultDataView200Response
type GetDefaultDataView200Response struct {
	DataViewId *string `json:"data_view_id,omitempty"`
}

// NewGetDefaultDataView200Response instantiates a new GetDefaultDataView200Response object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewGetDefaultDataView200Response() *GetDefaultDataView200Response {
	this := GetDefaultDataView200Response{}
	return &this
}

// NewGetDefaultDataView200ResponseWithDefaults instantiates a new GetDefaultDataView200Response object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewGetDefaultDataView200ResponseWithDefaults() *GetDefaultDataView200Response {
	this := GetDefaultDataView200Response{}
	return &this
}

// GetDataViewId returns the DataViewId field value if set, zero value otherwise.
func (o *GetDefaultDataView200Response) GetDataViewId() string {
	if o == nil || IsNil(o.DataViewId) {
		var ret string
		return ret
	}
	return *o.DataViewId
}

// GetDataViewIdOk returns a tuple with the DataViewId field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *GetDefaultDataView200Response) GetDataViewIdOk() (*string, bool) {
	if o == nil || IsNil(o.DataViewId) {
		return nil, false
	}
	return o.DataViewId, true
}

// HasDataViewId returns a boolean if a field has been set.
func (o *GetDefaultDataView200Response) HasDataViewId() bool {
	if o != nil && !IsNil(o.DataViewId) {
		return true
	}

	return false
}

// SetDataViewId gets a reference to the given string and assigns it to the DataViewId field.
func (o *GetDefaultDataView200Response) SetDataViewId(v string) {
	o.DataViewId = &v
}

func (o GetDefaultDataView200Response) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o GetDefaultDataView200Response) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	if !IsNil(o.DataViewId) {
		toSerialize["data_view_id"] = o.DataViewId
	}
	return toSerialize, nil
}

type NullableGetDefaultDataView200Response struct {
	value *GetDefaultDataView200Response
	isSet bool
}

func (v NullableGetDefaultDataView200Response) Get() *GetDefaultDataView200Response {
	return v.value
}

func (v *NullableGetDefaultDataView200Response) Set(val *GetDefaultDataView200Response) {
	v.value = val
	v.isSet = true
}

func (v NullableGetDefaultDataView200Response) IsSet() bool {
	return v.isSet
}

func (v *NullableGetDefaultDataView200Response) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableGetDefaultDataView200Response(val *GetDefaultDataView200Response) *NullableGetDefaultDataView200Response {
	return &NullableGetDefaultDataView200Response{value: val, isSet: true}
}

func (v NullableGetDefaultDataView200Response) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableGetDefaultDataView200Response) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
