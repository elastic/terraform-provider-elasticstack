/*
Alerting

OpenAPI schema for alerting endpoints

API version: 0.1
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package alerting

import (
	"encoding/json"
	"time"
)

// checks if the LegacyGetAlertingHealth200ResponseAlertingFrameworkHealthReadHealth type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &LegacyGetAlertingHealth200ResponseAlertingFrameworkHealthReadHealth{}

// LegacyGetAlertingHealth200ResponseAlertingFrameworkHealthReadHealth The timestamp and status of the alert reading events.
type LegacyGetAlertingHealth200ResponseAlertingFrameworkHealthReadHealth struct {
	Status    *string    `json:"status,omitempty"`
	Timestamp *time.Time `json:"timestamp,omitempty"`
}

// NewLegacyGetAlertingHealth200ResponseAlertingFrameworkHealthReadHealth instantiates a new LegacyGetAlertingHealth200ResponseAlertingFrameworkHealthReadHealth object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewLegacyGetAlertingHealth200ResponseAlertingFrameworkHealthReadHealth() *LegacyGetAlertingHealth200ResponseAlertingFrameworkHealthReadHealth {
	this := LegacyGetAlertingHealth200ResponseAlertingFrameworkHealthReadHealth{}
	return &this
}

// NewLegacyGetAlertingHealth200ResponseAlertingFrameworkHealthReadHealthWithDefaults instantiates a new LegacyGetAlertingHealth200ResponseAlertingFrameworkHealthReadHealth object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewLegacyGetAlertingHealth200ResponseAlertingFrameworkHealthReadHealthWithDefaults() *LegacyGetAlertingHealth200ResponseAlertingFrameworkHealthReadHealth {
	this := LegacyGetAlertingHealth200ResponseAlertingFrameworkHealthReadHealth{}
	return &this
}

// GetStatus returns the Status field value if set, zero value otherwise.
func (o *LegacyGetAlertingHealth200ResponseAlertingFrameworkHealthReadHealth) GetStatus() string {
	if o == nil || IsNil(o.Status) {
		var ret string
		return ret
	}
	return *o.Status
}

// GetStatusOk returns a tuple with the Status field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *LegacyGetAlertingHealth200ResponseAlertingFrameworkHealthReadHealth) GetStatusOk() (*string, bool) {
	if o == nil || IsNil(o.Status) {
		return nil, false
	}
	return o.Status, true
}

// HasStatus returns a boolean if a field has been set.
func (o *LegacyGetAlertingHealth200ResponseAlertingFrameworkHealthReadHealth) HasStatus() bool {
	if o != nil && !IsNil(o.Status) {
		return true
	}

	return false
}

// SetStatus gets a reference to the given string and assigns it to the Status field.
func (o *LegacyGetAlertingHealth200ResponseAlertingFrameworkHealthReadHealth) SetStatus(v string) {
	o.Status = &v
}

// GetTimestamp returns the Timestamp field value if set, zero value otherwise.
func (o *LegacyGetAlertingHealth200ResponseAlertingFrameworkHealthReadHealth) GetTimestamp() time.Time {
	if o == nil || IsNil(o.Timestamp) {
		var ret time.Time
		return ret
	}
	return *o.Timestamp
}

// GetTimestampOk returns a tuple with the Timestamp field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *LegacyGetAlertingHealth200ResponseAlertingFrameworkHealthReadHealth) GetTimestampOk() (*time.Time, bool) {
	if o == nil || IsNil(o.Timestamp) {
		return nil, false
	}
	return o.Timestamp, true
}

// HasTimestamp returns a boolean if a field has been set.
func (o *LegacyGetAlertingHealth200ResponseAlertingFrameworkHealthReadHealth) HasTimestamp() bool {
	if o != nil && !IsNil(o.Timestamp) {
		return true
	}

	return false
}

// SetTimestamp gets a reference to the given time.Time and assigns it to the Timestamp field.
func (o *LegacyGetAlertingHealth200ResponseAlertingFrameworkHealthReadHealth) SetTimestamp(v time.Time) {
	o.Timestamp = &v
}

func (o LegacyGetAlertingHealth200ResponseAlertingFrameworkHealthReadHealth) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o LegacyGetAlertingHealth200ResponseAlertingFrameworkHealthReadHealth) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	if !IsNil(o.Status) {
		toSerialize["status"] = o.Status
	}
	if !IsNil(o.Timestamp) {
		toSerialize["timestamp"] = o.Timestamp
	}
	return toSerialize, nil
}

type NullableLegacyGetAlertingHealth200ResponseAlertingFrameworkHealthReadHealth struct {
	value *LegacyGetAlertingHealth200ResponseAlertingFrameworkHealthReadHealth
	isSet bool
}

func (v NullableLegacyGetAlertingHealth200ResponseAlertingFrameworkHealthReadHealth) Get() *LegacyGetAlertingHealth200ResponseAlertingFrameworkHealthReadHealth {
	return v.value
}

func (v *NullableLegacyGetAlertingHealth200ResponseAlertingFrameworkHealthReadHealth) Set(val *LegacyGetAlertingHealth200ResponseAlertingFrameworkHealthReadHealth) {
	v.value = val
	v.isSet = true
}

func (v NullableLegacyGetAlertingHealth200ResponseAlertingFrameworkHealthReadHealth) IsSet() bool {
	return v.isSet
}

func (v *NullableLegacyGetAlertingHealth200ResponseAlertingFrameworkHealthReadHealth) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableLegacyGetAlertingHealth200ResponseAlertingFrameworkHealthReadHealth(val *LegacyGetAlertingHealth200ResponseAlertingFrameworkHealthReadHealth) *NullableLegacyGetAlertingHealth200ResponseAlertingFrameworkHealthReadHealth {
	return &NullableLegacyGetAlertingHealth200ResponseAlertingFrameworkHealthReadHealth{value: val, isSet: true}
}

func (v NullableLegacyGetAlertingHealth200ResponseAlertingFrameworkHealthReadHealth) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableLegacyGetAlertingHealth200ResponseAlertingFrameworkHealthReadHealth) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
