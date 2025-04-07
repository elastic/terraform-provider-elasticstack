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

// checks if the PostAlertingMaintenanceWindowRequestSchedule type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &PostAlertingMaintenanceWindowRequestSchedule{}

// PostAlertingMaintenanceWindowRequestSchedule struct for PostAlertingMaintenanceWindowRequestSchedule
type PostAlertingMaintenanceWindowRequestSchedule struct {
	Custom               PostAlertingMaintenanceWindowRequestScheduleCustom `json:"custom"`
	AdditionalProperties map[string]interface{}
}

type _PostAlertingMaintenanceWindowRequestSchedule PostAlertingMaintenanceWindowRequestSchedule

// NewPostAlertingMaintenanceWindowRequestSchedule instantiates a new PostAlertingMaintenanceWindowRequestSchedule object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewPostAlertingMaintenanceWindowRequestSchedule(custom PostAlertingMaintenanceWindowRequestScheduleCustom) *PostAlertingMaintenanceWindowRequestSchedule {
	this := PostAlertingMaintenanceWindowRequestSchedule{}
	this.Custom = custom
	return &this
}

// NewPostAlertingMaintenanceWindowRequestScheduleWithDefaults instantiates a new PostAlertingMaintenanceWindowRequestSchedule object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewPostAlertingMaintenanceWindowRequestScheduleWithDefaults() *PostAlertingMaintenanceWindowRequestSchedule {
	this := PostAlertingMaintenanceWindowRequestSchedule{}
	return &this
}

// GetCustom returns the Custom field value
func (o *PostAlertingMaintenanceWindowRequestSchedule) GetCustom() PostAlertingMaintenanceWindowRequestScheduleCustom {
	if o == nil {
		var ret PostAlertingMaintenanceWindowRequestScheduleCustom
		return ret
	}

	return o.Custom
}

// GetCustomOk returns a tuple with the Custom field value
// and a boolean to check if the value has been set.
func (o *PostAlertingMaintenanceWindowRequestSchedule) GetCustomOk() (*PostAlertingMaintenanceWindowRequestScheduleCustom, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Custom, true
}

// SetCustom sets field value
func (o *PostAlertingMaintenanceWindowRequestSchedule) SetCustom(v PostAlertingMaintenanceWindowRequestScheduleCustom) {
	o.Custom = v
}

func (o PostAlertingMaintenanceWindowRequestSchedule) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o PostAlertingMaintenanceWindowRequestSchedule) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["custom"] = o.Custom

	for key, value := range o.AdditionalProperties {
		toSerialize[key] = value
	}

	return toSerialize, nil
}

func (o *PostAlertingMaintenanceWindowRequestSchedule) UnmarshalJSON(bytes []byte) (err error) {
	varPostAlertingMaintenanceWindowRequestSchedule := _PostAlertingMaintenanceWindowRequestSchedule{}

	err = json.Unmarshal(bytes, &varPostAlertingMaintenanceWindowRequestSchedule)

	if err != nil {
		return err
	}

	*o = PostAlertingMaintenanceWindowRequestSchedule(varPostAlertingMaintenanceWindowRequestSchedule)

	additionalProperties := make(map[string]interface{})

	if err = json.Unmarshal(bytes, &additionalProperties); err == nil {
		delete(additionalProperties, "custom")
		o.AdditionalProperties = additionalProperties
	}

	return err
}

type NullablePostAlertingMaintenanceWindowRequestSchedule struct {
	value *PostAlertingMaintenanceWindowRequestSchedule
	isSet bool
}

func (v NullablePostAlertingMaintenanceWindowRequestSchedule) Get() *PostAlertingMaintenanceWindowRequestSchedule {
	return v.value
}

func (v *NullablePostAlertingMaintenanceWindowRequestSchedule) Set(val *PostAlertingMaintenanceWindowRequestSchedule) {
	v.value = val
	v.isSet = true
}

func (v NullablePostAlertingMaintenanceWindowRequestSchedule) IsSet() bool {
	return v.isSet
}

func (v *NullablePostAlertingMaintenanceWindowRequestSchedule) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullablePostAlertingMaintenanceWindowRequestSchedule(val *PostAlertingMaintenanceWindowRequestSchedule) *NullablePostAlertingMaintenanceWindowRequestSchedule {
	return &NullablePostAlertingMaintenanceWindowRequestSchedule{value: val, isSet: true}
}

func (v NullablePostAlertingMaintenanceWindowRequestSchedule) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullablePostAlertingMaintenanceWindowRequestSchedule) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
