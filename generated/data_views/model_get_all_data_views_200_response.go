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

// checks if the GetAllDataViews200Response type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &GetAllDataViews200Response{}

// GetAllDataViews200Response struct for GetAllDataViews200Response
type GetAllDataViews200Response struct {
	DataView []GetAllDataViews200ResponseDataViewInner `json:"data_view,omitempty"`
}

// NewGetAllDataViews200Response instantiates a new GetAllDataViews200Response object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewGetAllDataViews200Response() *GetAllDataViews200Response {
	this := GetAllDataViews200Response{}
	return &this
}

// NewGetAllDataViews200ResponseWithDefaults instantiates a new GetAllDataViews200Response object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewGetAllDataViews200ResponseWithDefaults() *GetAllDataViews200Response {
	this := GetAllDataViews200Response{}
	return &this
}

// GetDataView returns the DataView field value if set, zero value otherwise.
func (o *GetAllDataViews200Response) GetDataView() []GetAllDataViews200ResponseDataViewInner {
	if o == nil || IsNil(o.DataView) {
		var ret []GetAllDataViews200ResponseDataViewInner
		return ret
	}
	return o.DataView
}

// GetDataViewOk returns a tuple with the DataView field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *GetAllDataViews200Response) GetDataViewOk() ([]GetAllDataViews200ResponseDataViewInner, bool) {
	if o == nil || IsNil(o.DataView) {
		return nil, false
	}
	return o.DataView, true
}

// HasDataView returns a boolean if a field has been set.
func (o *GetAllDataViews200Response) HasDataView() bool {
	if o != nil && !IsNil(o.DataView) {
		return true
	}

	return false
}

// SetDataView gets a reference to the given []GetAllDataViews200ResponseDataViewInner and assigns it to the DataView field.
func (o *GetAllDataViews200Response) SetDataView(v []GetAllDataViews200ResponseDataViewInner) {
	o.DataView = v
}

func (o GetAllDataViews200Response) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o GetAllDataViews200Response) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	if !IsNil(o.DataView) {
		toSerialize["data_view"] = o.DataView
	}
	return toSerialize, nil
}

type NullableGetAllDataViews200Response struct {
	value *GetAllDataViews200Response
	isSet bool
}

func (v NullableGetAllDataViews200Response) Get() *GetAllDataViews200Response {
	return v.value
}

func (v *NullableGetAllDataViews200Response) Set(val *GetAllDataViews200Response) {
	v.value = val
	v.isSet = true
}

func (v NullableGetAllDataViews200Response) IsSet() bool {
	return v.isSet
}

func (v *NullableGetAllDataViews200Response) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableGetAllDataViews200Response(val *GetAllDataViews200Response) *NullableGetAllDataViews200Response {
	return &NullableGetAllDataViews200Response{value: val, isSet: true}
}

func (v NullableGetAllDataViews200Response) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableGetAllDataViews200Response) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
