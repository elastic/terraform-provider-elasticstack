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

// checks if the UpdateDataViewRequestObjectDataView type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &UpdateDataViewRequestObjectDataView{}

// UpdateDataViewRequestObjectDataView The data view properties you want to update. Only the specified properties are updated in the data view. Unspecified fields stay as they are persisted.
type UpdateDataViewRequestObjectDataView struct {
	// Allows the data view saved object to exist before the data is available.
	AllowNoIndex interface{} `json:"allowNoIndex,omitempty"`
	// A map of field formats by field name.
	FieldFormats interface{} `json:"fieldFormats,omitempty"`
	Fields       interface{} `json:"fields,omitempty"`
	Name         interface{} `json:"name,omitempty"`
	// A map of runtime field definitions by field name.
	RuntimeFieldMap interface{} `json:"runtimeFieldMap,omitempty"`
	// The array of field names you want to filter out in Discover.
	SourceFilters interface{} `json:"sourceFilters,omitempty"`
	// The timestamp field name, which you use for time-based data views.
	TimeFieldName interface{} `json:"timeFieldName,omitempty"`
	// Comma-separated list of data streams, indices, and aliases that you want to search. Supports wildcards (`*`).
	Title interface{} `json:"title,omitempty"`
	// When set to `rollup`, identifies the rollup data views.
	Type interface{} `json:"type,omitempty"`
	// When you use rollup indices, contains the field list for the rollup data view API endpoints.
	TypeMeta interface{} `json:"typeMeta,omitempty"`
}

// NewUpdateDataViewRequestObjectDataView instantiates a new UpdateDataViewRequestObjectDataView object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewUpdateDataViewRequestObjectDataView() *UpdateDataViewRequestObjectDataView {
	this := UpdateDataViewRequestObjectDataView{}
	return &this
}

// NewUpdateDataViewRequestObjectDataViewWithDefaults instantiates a new UpdateDataViewRequestObjectDataView object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewUpdateDataViewRequestObjectDataViewWithDefaults() *UpdateDataViewRequestObjectDataView {
	this := UpdateDataViewRequestObjectDataView{}
	return &this
}

// GetAllowNoIndex returns the AllowNoIndex field value if set, zero value otherwise (both if not set or set to explicit null).
func (o *UpdateDataViewRequestObjectDataView) GetAllowNoIndex() interface{} {
	if o == nil {
		var ret interface{}
		return ret
	}
	return o.AllowNoIndex
}

// GetAllowNoIndexOk returns a tuple with the AllowNoIndex field value if set, nil otherwise
// and a boolean to check if the value has been set.
// NOTE: If the value is an explicit nil, `nil, true` will be returned
func (o *UpdateDataViewRequestObjectDataView) GetAllowNoIndexOk() (*interface{}, bool) {
	if o == nil || IsNil(o.AllowNoIndex) {
		return nil, false
	}
	return &o.AllowNoIndex, true
}

// HasAllowNoIndex returns a boolean if a field has been set.
func (o *UpdateDataViewRequestObjectDataView) HasAllowNoIndex() bool {
	if o != nil && IsNil(o.AllowNoIndex) {
		return true
	}

	return false
}

// SetAllowNoIndex gets a reference to the given interface{} and assigns it to the AllowNoIndex field.
func (o *UpdateDataViewRequestObjectDataView) SetAllowNoIndex(v interface{}) {
	o.AllowNoIndex = v
}

// GetFieldFormats returns the FieldFormats field value if set, zero value otherwise (both if not set or set to explicit null).
func (o *UpdateDataViewRequestObjectDataView) GetFieldFormats() interface{} {
	if o == nil {
		var ret interface{}
		return ret
	}
	return o.FieldFormats
}

// GetFieldFormatsOk returns a tuple with the FieldFormats field value if set, nil otherwise
// and a boolean to check if the value has been set.
// NOTE: If the value is an explicit nil, `nil, true` will be returned
func (o *UpdateDataViewRequestObjectDataView) GetFieldFormatsOk() (*interface{}, bool) {
	if o == nil || IsNil(o.FieldFormats) {
		return nil, false
	}
	return &o.FieldFormats, true
}

// HasFieldFormats returns a boolean if a field has been set.
func (o *UpdateDataViewRequestObjectDataView) HasFieldFormats() bool {
	if o != nil && IsNil(o.FieldFormats) {
		return true
	}

	return false
}

// SetFieldFormats gets a reference to the given interface{} and assigns it to the FieldFormats field.
func (o *UpdateDataViewRequestObjectDataView) SetFieldFormats(v interface{}) {
	o.FieldFormats = v
}

// GetFields returns the Fields field value if set, zero value otherwise (both if not set or set to explicit null).
func (o *UpdateDataViewRequestObjectDataView) GetFields() interface{} {
	if o == nil {
		var ret interface{}
		return ret
	}
	return o.Fields
}

// GetFieldsOk returns a tuple with the Fields field value if set, nil otherwise
// and a boolean to check if the value has been set.
// NOTE: If the value is an explicit nil, `nil, true` will be returned
func (o *UpdateDataViewRequestObjectDataView) GetFieldsOk() (*interface{}, bool) {
	if o == nil || IsNil(o.Fields) {
		return nil, false
	}
	return &o.Fields, true
}

// HasFields returns a boolean if a field has been set.
func (o *UpdateDataViewRequestObjectDataView) HasFields() bool {
	if o != nil && IsNil(o.Fields) {
		return true
	}

	return false
}

// SetFields gets a reference to the given interface{} and assigns it to the Fields field.
func (o *UpdateDataViewRequestObjectDataView) SetFields(v interface{}) {
	o.Fields = v
}

// GetName returns the Name field value if set, zero value otherwise (both if not set or set to explicit null).
func (o *UpdateDataViewRequestObjectDataView) GetName() interface{} {
	if o == nil {
		var ret interface{}
		return ret
	}
	return o.Name
}

// GetNameOk returns a tuple with the Name field value if set, nil otherwise
// and a boolean to check if the value has been set.
// NOTE: If the value is an explicit nil, `nil, true` will be returned
func (o *UpdateDataViewRequestObjectDataView) GetNameOk() (*interface{}, bool) {
	if o == nil || IsNil(o.Name) {
		return nil, false
	}
	return &o.Name, true
}

// HasName returns a boolean if a field has been set.
func (o *UpdateDataViewRequestObjectDataView) HasName() bool {
	if o != nil && IsNil(o.Name) {
		return true
	}

	return false
}

// SetName gets a reference to the given interface{} and assigns it to the Name field.
func (o *UpdateDataViewRequestObjectDataView) SetName(v interface{}) {
	o.Name = v
}

// GetRuntimeFieldMap returns the RuntimeFieldMap field value if set, zero value otherwise (both if not set or set to explicit null).
func (o *UpdateDataViewRequestObjectDataView) GetRuntimeFieldMap() interface{} {
	if o == nil {
		var ret interface{}
		return ret
	}
	return o.RuntimeFieldMap
}

// GetRuntimeFieldMapOk returns a tuple with the RuntimeFieldMap field value if set, nil otherwise
// and a boolean to check if the value has been set.
// NOTE: If the value is an explicit nil, `nil, true` will be returned
func (o *UpdateDataViewRequestObjectDataView) GetRuntimeFieldMapOk() (*interface{}, bool) {
	if o == nil || IsNil(o.RuntimeFieldMap) {
		return nil, false
	}
	return &o.RuntimeFieldMap, true
}

// HasRuntimeFieldMap returns a boolean if a field has been set.
func (o *UpdateDataViewRequestObjectDataView) HasRuntimeFieldMap() bool {
	if o != nil && IsNil(o.RuntimeFieldMap) {
		return true
	}

	return false
}

// SetRuntimeFieldMap gets a reference to the given interface{} and assigns it to the RuntimeFieldMap field.
func (o *UpdateDataViewRequestObjectDataView) SetRuntimeFieldMap(v interface{}) {
	o.RuntimeFieldMap = v
}

// GetSourceFilters returns the SourceFilters field value if set, zero value otherwise (both if not set or set to explicit null).
func (o *UpdateDataViewRequestObjectDataView) GetSourceFilters() interface{} {
	if o == nil {
		var ret interface{}
		return ret
	}
	return o.SourceFilters
}

// GetSourceFiltersOk returns a tuple with the SourceFilters field value if set, nil otherwise
// and a boolean to check if the value has been set.
// NOTE: If the value is an explicit nil, `nil, true` will be returned
func (o *UpdateDataViewRequestObjectDataView) GetSourceFiltersOk() (*interface{}, bool) {
	if o == nil || IsNil(o.SourceFilters) {
		return nil, false
	}
	return &o.SourceFilters, true
}

// HasSourceFilters returns a boolean if a field has been set.
func (o *UpdateDataViewRequestObjectDataView) HasSourceFilters() bool {
	if o != nil && IsNil(o.SourceFilters) {
		return true
	}

	return false
}

// SetSourceFilters gets a reference to the given interface{} and assigns it to the SourceFilters field.
func (o *UpdateDataViewRequestObjectDataView) SetSourceFilters(v interface{}) {
	o.SourceFilters = v
}

// GetTimeFieldName returns the TimeFieldName field value if set, zero value otherwise (both if not set or set to explicit null).
func (o *UpdateDataViewRequestObjectDataView) GetTimeFieldName() interface{} {
	if o == nil {
		var ret interface{}
		return ret
	}
	return o.TimeFieldName
}

// GetTimeFieldNameOk returns a tuple with the TimeFieldName field value if set, nil otherwise
// and a boolean to check if the value has been set.
// NOTE: If the value is an explicit nil, `nil, true` will be returned
func (o *UpdateDataViewRequestObjectDataView) GetTimeFieldNameOk() (*interface{}, bool) {
	if o == nil || IsNil(o.TimeFieldName) {
		return nil, false
	}
	return &o.TimeFieldName, true
}

// HasTimeFieldName returns a boolean if a field has been set.
func (o *UpdateDataViewRequestObjectDataView) HasTimeFieldName() bool {
	if o != nil && IsNil(o.TimeFieldName) {
		return true
	}

	return false
}

// SetTimeFieldName gets a reference to the given interface{} and assigns it to the TimeFieldName field.
func (o *UpdateDataViewRequestObjectDataView) SetTimeFieldName(v interface{}) {
	o.TimeFieldName = v
}

// GetTitle returns the Title field value if set, zero value otherwise (both if not set or set to explicit null).
func (o *UpdateDataViewRequestObjectDataView) GetTitle() interface{} {
	if o == nil {
		var ret interface{}
		return ret
	}
	return o.Title
}

// GetTitleOk returns a tuple with the Title field value if set, nil otherwise
// and a boolean to check if the value has been set.
// NOTE: If the value is an explicit nil, `nil, true` will be returned
func (o *UpdateDataViewRequestObjectDataView) GetTitleOk() (*interface{}, bool) {
	if o == nil || IsNil(o.Title) {
		return nil, false
	}
	return &o.Title, true
}

// HasTitle returns a boolean if a field has been set.
func (o *UpdateDataViewRequestObjectDataView) HasTitle() bool {
	if o != nil && IsNil(o.Title) {
		return true
	}

	return false
}

// SetTitle gets a reference to the given interface{} and assigns it to the Title field.
func (o *UpdateDataViewRequestObjectDataView) SetTitle(v interface{}) {
	o.Title = v
}

// GetType returns the Type field value if set, zero value otherwise (both if not set or set to explicit null).
func (o *UpdateDataViewRequestObjectDataView) GetType() interface{} {
	if o == nil {
		var ret interface{}
		return ret
	}
	return o.Type
}

// GetTypeOk returns a tuple with the Type field value if set, nil otherwise
// and a boolean to check if the value has been set.
// NOTE: If the value is an explicit nil, `nil, true` will be returned
func (o *UpdateDataViewRequestObjectDataView) GetTypeOk() (*interface{}, bool) {
	if o == nil || IsNil(o.Type) {
		return nil, false
	}
	return &o.Type, true
}

// HasType returns a boolean if a field has been set.
func (o *UpdateDataViewRequestObjectDataView) HasType() bool {
	if o != nil && IsNil(o.Type) {
		return true
	}

	return false
}

// SetType gets a reference to the given interface{} and assigns it to the Type field.
func (o *UpdateDataViewRequestObjectDataView) SetType(v interface{}) {
	o.Type = v
}

// GetTypeMeta returns the TypeMeta field value if set, zero value otherwise (both if not set or set to explicit null).
func (o *UpdateDataViewRequestObjectDataView) GetTypeMeta() interface{} {
	if o == nil {
		var ret interface{}
		return ret
	}
	return o.TypeMeta
}

// GetTypeMetaOk returns a tuple with the TypeMeta field value if set, nil otherwise
// and a boolean to check if the value has been set.
// NOTE: If the value is an explicit nil, `nil, true` will be returned
func (o *UpdateDataViewRequestObjectDataView) GetTypeMetaOk() (*interface{}, bool) {
	if o == nil || IsNil(o.TypeMeta) {
		return nil, false
	}
	return &o.TypeMeta, true
}

// HasTypeMeta returns a boolean if a field has been set.
func (o *UpdateDataViewRequestObjectDataView) HasTypeMeta() bool {
	if o != nil && IsNil(o.TypeMeta) {
		return true
	}

	return false
}

// SetTypeMeta gets a reference to the given interface{} and assigns it to the TypeMeta field.
func (o *UpdateDataViewRequestObjectDataView) SetTypeMeta(v interface{}) {
	o.TypeMeta = v
}

func (o UpdateDataViewRequestObjectDataView) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o UpdateDataViewRequestObjectDataView) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	if o.AllowNoIndex != nil {
		toSerialize["allowNoIndex"] = o.AllowNoIndex
	}
	if o.FieldFormats != nil {
		toSerialize["fieldFormats"] = o.FieldFormats
	}
	if o.Fields != nil {
		toSerialize["fields"] = o.Fields
	}
	if o.Name != nil {
		toSerialize["name"] = o.Name
	}
	if o.RuntimeFieldMap != nil {
		toSerialize["runtimeFieldMap"] = o.RuntimeFieldMap
	}
	if o.SourceFilters != nil {
		toSerialize["sourceFilters"] = o.SourceFilters
	}
	if o.TimeFieldName != nil {
		toSerialize["timeFieldName"] = o.TimeFieldName
	}
	if o.Title != nil {
		toSerialize["title"] = o.Title
	}
	if o.Type != nil {
		toSerialize["type"] = o.Type
	}
	if o.TypeMeta != nil {
		toSerialize["typeMeta"] = o.TypeMeta
	}
	return toSerialize, nil
}

type NullableUpdateDataViewRequestObjectDataView struct {
	value *UpdateDataViewRequestObjectDataView
	isSet bool
}

func (v NullableUpdateDataViewRequestObjectDataView) Get() *UpdateDataViewRequestObjectDataView {
	return v.value
}

func (v *NullableUpdateDataViewRequestObjectDataView) Set(val *UpdateDataViewRequestObjectDataView) {
	v.value = val
	v.isSet = true
}

func (v NullableUpdateDataViewRequestObjectDataView) IsSet() bool {
	return v.isSet
}

func (v *NullableUpdateDataViewRequestObjectDataView) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableUpdateDataViewRequestObjectDataView(val *UpdateDataViewRequestObjectDataView) *NullableUpdateDataViewRequestObjectDataView {
	return &NullableUpdateDataViewRequestObjectDataView{value: val, isSet: true}
}

func (v NullableUpdateDataViewRequestObjectDataView) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableUpdateDataViewRequestObjectDataView) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}