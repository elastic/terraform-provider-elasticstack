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

// checks if the PostAlertingMaintenanceWindowRequestScopeAlertingQuery type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &PostAlertingMaintenanceWindowRequestScopeAlertingQuery{}

// PostAlertingMaintenanceWindowRequestScopeAlertingQuery struct for PostAlertingMaintenanceWindowRequestScopeAlertingQuery
type PostAlertingMaintenanceWindowRequestScopeAlertingQuery struct {
	// A filter written in Kibana Query Language (KQL).
	Kql                  interface{} `json:"kql"`
	AdditionalProperties map[string]interface{}
}

type _PostAlertingMaintenanceWindowRequestScopeAlertingQuery PostAlertingMaintenanceWindowRequestScopeAlertingQuery

// NewPostAlertingMaintenanceWindowRequestScopeAlertingQuery instantiates a new PostAlertingMaintenanceWindowRequestScopeAlertingQuery object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewPostAlertingMaintenanceWindowRequestScopeAlertingQuery(kql interface{}) *PostAlertingMaintenanceWindowRequestScopeAlertingQuery {
	this := PostAlertingMaintenanceWindowRequestScopeAlertingQuery{}
	this.Kql = kql
	return &this
}

// NewPostAlertingMaintenanceWindowRequestScopeAlertingQueryWithDefaults instantiates a new PostAlertingMaintenanceWindowRequestScopeAlertingQuery object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewPostAlertingMaintenanceWindowRequestScopeAlertingQueryWithDefaults() *PostAlertingMaintenanceWindowRequestScopeAlertingQuery {
	this := PostAlertingMaintenanceWindowRequestScopeAlertingQuery{}
	return &this
}

// GetKql returns the Kql field value
// If the value is explicit nil, the zero value for interface{} will be returned
func (o *PostAlertingMaintenanceWindowRequestScopeAlertingQuery) GetKql() interface{} {
	if o == nil {
		var ret interface{}
		return ret
	}

	return o.Kql
}

// GetKqlOk returns a tuple with the Kql field value
// and a boolean to check if the value has been set.
// NOTE: If the value is an explicit nil, `nil, true` will be returned
func (o *PostAlertingMaintenanceWindowRequestScopeAlertingQuery) GetKqlOk() (*interface{}, bool) {
	if o == nil || IsNil(o.Kql) {
		return nil, false
	}
	return &o.Kql, true
}

// SetKql sets field value
func (o *PostAlertingMaintenanceWindowRequestScopeAlertingQuery) SetKql(v interface{}) {
	o.Kql = v
}

func (o PostAlertingMaintenanceWindowRequestScopeAlertingQuery) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o PostAlertingMaintenanceWindowRequestScopeAlertingQuery) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	if o.Kql != nil {
		toSerialize["kql"] = o.Kql
	}

	for key, value := range o.AdditionalProperties {
		toSerialize[key] = value
	}

	return toSerialize, nil
}

func (o *PostAlertingMaintenanceWindowRequestScopeAlertingQuery) UnmarshalJSON(bytes []byte) (err error) {
	varPostAlertingMaintenanceWindowRequestScopeAlertingQuery := _PostAlertingMaintenanceWindowRequestScopeAlertingQuery{}

	err = json.Unmarshal(bytes, &varPostAlertingMaintenanceWindowRequestScopeAlertingQuery)

	if err != nil {
		return err
	}

	*o = PostAlertingMaintenanceWindowRequestScopeAlertingQuery(varPostAlertingMaintenanceWindowRequestScopeAlertingQuery)

	additionalProperties := make(map[string]interface{})

	if err = json.Unmarshal(bytes, &additionalProperties); err == nil {
		delete(additionalProperties, "kql")
		o.AdditionalProperties = additionalProperties
	}

	return err
}

type NullablePostAlertingMaintenanceWindowRequestScopeAlertingQuery struct {
	value *PostAlertingMaintenanceWindowRequestScopeAlertingQuery
	isSet bool
}

func (v NullablePostAlertingMaintenanceWindowRequestScopeAlertingQuery) Get() *PostAlertingMaintenanceWindowRequestScopeAlertingQuery {
	return v.value
}

func (v *NullablePostAlertingMaintenanceWindowRequestScopeAlertingQuery) Set(val *PostAlertingMaintenanceWindowRequestScopeAlertingQuery) {
	v.value = val
	v.isSet = true
}

func (v NullablePostAlertingMaintenanceWindowRequestScopeAlertingQuery) IsSet() bool {
	return v.isSet
}

func (v *NullablePostAlertingMaintenanceWindowRequestScopeAlertingQuery) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullablePostAlertingMaintenanceWindowRequestScopeAlertingQuery(val *PostAlertingMaintenanceWindowRequestScopeAlertingQuery) *NullablePostAlertingMaintenanceWindowRequestScopeAlertingQuery {
	return &NullablePostAlertingMaintenanceWindowRequestScopeAlertingQuery{value: val, isSet: true}
}

func (v NullablePostAlertingMaintenanceWindowRequestScopeAlertingQuery) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullablePostAlertingMaintenanceWindowRequestScopeAlertingQuery) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
