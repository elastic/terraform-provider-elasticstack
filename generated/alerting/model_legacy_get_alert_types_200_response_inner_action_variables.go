/*
Alerting

OpenAPI schema for alerting endpoints

API version: 0.1
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package alerting

import (
	"encoding/json"
)

// checks if the LegacyGetAlertTypes200ResponseInnerActionVariables type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &LegacyGetAlertTypes200ResponseInnerActionVariables{}

// LegacyGetAlertTypes200ResponseInnerActionVariables A list of action variables that the alert type makes available via context and state in action parameter templates, and a short human readable description. The Alert UI will use this information to prompt users for these variables in action parameter editors.
type LegacyGetAlertTypes200ResponseInnerActionVariables struct {
	Context []LegacyGetAlertTypes200ResponseInnerActionVariablesContextInner `json:"context,omitempty"`
	Params  []GetRuleTypes200ResponseInnerActionVariablesParamsInner         `json:"params,omitempty"`
	State   []GetRuleTypes200ResponseInnerActionVariablesParamsInner         `json:"state,omitempty"`
}

// NewLegacyGetAlertTypes200ResponseInnerActionVariables instantiates a new LegacyGetAlertTypes200ResponseInnerActionVariables object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewLegacyGetAlertTypes200ResponseInnerActionVariables() *LegacyGetAlertTypes200ResponseInnerActionVariables {
	this := LegacyGetAlertTypes200ResponseInnerActionVariables{}
	return &this
}

// NewLegacyGetAlertTypes200ResponseInnerActionVariablesWithDefaults instantiates a new LegacyGetAlertTypes200ResponseInnerActionVariables object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewLegacyGetAlertTypes200ResponseInnerActionVariablesWithDefaults() *LegacyGetAlertTypes200ResponseInnerActionVariables {
	this := LegacyGetAlertTypes200ResponseInnerActionVariables{}
	return &this
}

// GetContext returns the Context field value if set, zero value otherwise.
func (o *LegacyGetAlertTypes200ResponseInnerActionVariables) GetContext() []LegacyGetAlertTypes200ResponseInnerActionVariablesContextInner {
	if o == nil || IsNil(o.Context) {
		var ret []LegacyGetAlertTypes200ResponseInnerActionVariablesContextInner
		return ret
	}
	return o.Context
}

// GetContextOk returns a tuple with the Context field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *LegacyGetAlertTypes200ResponseInnerActionVariables) GetContextOk() ([]LegacyGetAlertTypes200ResponseInnerActionVariablesContextInner, bool) {
	if o == nil || IsNil(o.Context) {
		return nil, false
	}
	return o.Context, true
}

// HasContext returns a boolean if a field has been set.
func (o *LegacyGetAlertTypes200ResponseInnerActionVariables) HasContext() bool {
	if o != nil && !IsNil(o.Context) {
		return true
	}

	return false
}

// SetContext gets a reference to the given []LegacyGetAlertTypes200ResponseInnerActionVariablesContextInner and assigns it to the Context field.
func (o *LegacyGetAlertTypes200ResponseInnerActionVariables) SetContext(v []LegacyGetAlertTypes200ResponseInnerActionVariablesContextInner) {
	o.Context = v
}

// GetParams returns the Params field value if set, zero value otherwise.
func (o *LegacyGetAlertTypes200ResponseInnerActionVariables) GetParams() []GetRuleTypes200ResponseInnerActionVariablesParamsInner {
	if o == nil || IsNil(o.Params) {
		var ret []GetRuleTypes200ResponseInnerActionVariablesParamsInner
		return ret
	}
	return o.Params
}

// GetParamsOk returns a tuple with the Params field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *LegacyGetAlertTypes200ResponseInnerActionVariables) GetParamsOk() ([]GetRuleTypes200ResponseInnerActionVariablesParamsInner, bool) {
	if o == nil || IsNil(o.Params) {
		return nil, false
	}
	return o.Params, true
}

// HasParams returns a boolean if a field has been set.
func (o *LegacyGetAlertTypes200ResponseInnerActionVariables) HasParams() bool {
	if o != nil && !IsNil(o.Params) {
		return true
	}

	return false
}

// SetParams gets a reference to the given []GetRuleTypes200ResponseInnerActionVariablesParamsInner and assigns it to the Params field.
func (o *LegacyGetAlertTypes200ResponseInnerActionVariables) SetParams(v []GetRuleTypes200ResponseInnerActionVariablesParamsInner) {
	o.Params = v
}

// GetState returns the State field value if set, zero value otherwise.
func (o *LegacyGetAlertTypes200ResponseInnerActionVariables) GetState() []GetRuleTypes200ResponseInnerActionVariablesParamsInner {
	if o == nil || IsNil(o.State) {
		var ret []GetRuleTypes200ResponseInnerActionVariablesParamsInner
		return ret
	}
	return o.State
}

// GetStateOk returns a tuple with the State field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *LegacyGetAlertTypes200ResponseInnerActionVariables) GetStateOk() ([]GetRuleTypes200ResponseInnerActionVariablesParamsInner, bool) {
	if o == nil || IsNil(o.State) {
		return nil, false
	}
	return o.State, true
}

// HasState returns a boolean if a field has been set.
func (o *LegacyGetAlertTypes200ResponseInnerActionVariables) HasState() bool {
	if o != nil && !IsNil(o.State) {
		return true
	}

	return false
}

// SetState gets a reference to the given []GetRuleTypes200ResponseInnerActionVariablesParamsInner and assigns it to the State field.
func (o *LegacyGetAlertTypes200ResponseInnerActionVariables) SetState(v []GetRuleTypes200ResponseInnerActionVariablesParamsInner) {
	o.State = v
}

func (o LegacyGetAlertTypes200ResponseInnerActionVariables) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o LegacyGetAlertTypes200ResponseInnerActionVariables) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	if !IsNil(o.Context) {
		toSerialize["context"] = o.Context
	}
	if !IsNil(o.Params) {
		toSerialize["params"] = o.Params
	}
	if !IsNil(o.State) {
		toSerialize["state"] = o.State
	}
	return toSerialize, nil
}

type NullableLegacyGetAlertTypes200ResponseInnerActionVariables struct {
	value *LegacyGetAlertTypes200ResponseInnerActionVariables
	isSet bool
}

func (v NullableLegacyGetAlertTypes200ResponseInnerActionVariables) Get() *LegacyGetAlertTypes200ResponseInnerActionVariables {
	return v.value
}

func (v *NullableLegacyGetAlertTypes200ResponseInnerActionVariables) Set(val *LegacyGetAlertTypes200ResponseInnerActionVariables) {
	v.value = val
	v.isSet = true
}

func (v NullableLegacyGetAlertTypes200ResponseInnerActionVariables) IsSet() bool {
	return v.isSet
}

func (v *NullableLegacyGetAlertTypes200ResponseInnerActionVariables) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableLegacyGetAlertTypes200ResponseInnerActionVariables(val *LegacyGetAlertTypes200ResponseInnerActionVariables) *NullableLegacyGetAlertTypes200ResponseInnerActionVariables {
	return &NullableLegacyGetAlertTypes200ResponseInnerActionVariables{value: val, isSet: true}
}

func (v NullableLegacyGetAlertTypes200ResponseInnerActionVariables) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableLegacyGetAlertTypes200ResponseInnerActionVariables) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
