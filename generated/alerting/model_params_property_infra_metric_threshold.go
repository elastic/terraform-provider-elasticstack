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

// checks if the ParamsPropertyInfraMetricThreshold type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &ParamsPropertyInfraMetricThreshold{}

// ParamsPropertyInfraMetricThreshold struct for ParamsPropertyInfraMetricThreshold
type ParamsPropertyInfraMetricThreshold struct {
	Criteria              []*OneOf       `json:"criteria,omitempty"`
	GroupBy               NullableString `json:"groupBy,omitempty"`
	FilterQuery           *string        `json:"filterQuery,omitempty"`
	SourceId              *string        `json:"sourceId,omitempty"`
	AlertOnNoData         *bool          `json:"alertOnNoData,omitempty"`
	AlertOnGroupDisappear *bool          `json:"alertOnGroupDisappear,omitempty"`
}

// NewParamsPropertyInfraMetricThreshold instantiates a new ParamsPropertyInfraMetricThreshold object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewParamsPropertyInfraMetricThreshold() *ParamsPropertyInfraMetricThreshold {
	this := ParamsPropertyInfraMetricThreshold{}
	return &this
}

// NewParamsPropertyInfraMetricThresholdWithDefaults instantiates a new ParamsPropertyInfraMetricThreshold object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewParamsPropertyInfraMetricThresholdWithDefaults() *ParamsPropertyInfraMetricThreshold {
	this := ParamsPropertyInfraMetricThreshold{}
	return &this
}

// GetCriteria returns the Criteria field value if set, zero value otherwise.
func (o *ParamsPropertyInfraMetricThreshold) GetCriteria() []*OneOf {
	if o == nil || IsNil(o.Criteria) {
		var ret []*OneOf
		return ret
	}
	return o.Criteria
}

// GetCriteriaOk returns a tuple with the Criteria field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ParamsPropertyInfraMetricThreshold) GetCriteriaOk() ([]*OneOf, bool) {
	if o == nil || IsNil(o.Criteria) {
		return nil, false
	}
	return o.Criteria, true
}

// HasCriteria returns a boolean if a field has been set.
func (o *ParamsPropertyInfraMetricThreshold) HasCriteria() bool {
	if o != nil && !IsNil(o.Criteria) {
		return true
	}

	return false
}

// SetCriteria gets a reference to the given []*OneOf and assigns it to the Criteria field.
func (o *ParamsPropertyInfraMetricThreshold) SetCriteria(v []*OneOf) {
	o.Criteria = v
}

// GetGroupBy returns the GroupBy field value if set, zero value otherwise (both if not set or set to explicit null).
func (o *ParamsPropertyInfraMetricThreshold) GetGroupBy() string {
	if o == nil || IsNil(o.GroupBy.Get()) {
		var ret string
		return ret
	}
	return *o.GroupBy.Get()
}

// GetGroupByOk returns a tuple with the GroupBy field value if set, nil otherwise
// and a boolean to check if the value has been set.
// NOTE: If the value is an explicit nil, `nil, true` will be returned
func (o *ParamsPropertyInfraMetricThreshold) GetGroupByOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return o.GroupBy.Get(), o.GroupBy.IsSet()
}

// HasGroupBy returns a boolean if a field has been set.
func (o *ParamsPropertyInfraMetricThreshold) HasGroupBy() bool {
	if o != nil && o.GroupBy.IsSet() {
		return true
	}

	return false
}

// SetGroupBy gets a reference to the given NullableString and assigns it to the GroupBy field.
func (o *ParamsPropertyInfraMetricThreshold) SetGroupBy(v string) {
	o.GroupBy.Set(&v)
}

// SetGroupByNil sets the value for GroupBy to be an explicit nil
func (o *ParamsPropertyInfraMetricThreshold) SetGroupByNil() {
	o.GroupBy.Set(nil)
}

// UnsetGroupBy ensures that no value is present for GroupBy, not even an explicit nil
func (o *ParamsPropertyInfraMetricThreshold) UnsetGroupBy() {
	o.GroupBy.Unset()
}

// GetFilterQuery returns the FilterQuery field value if set, zero value otherwise.
func (o *ParamsPropertyInfraMetricThreshold) GetFilterQuery() string {
	if o == nil || IsNil(o.FilterQuery) {
		var ret string
		return ret
	}
	return *o.FilterQuery
}

// GetFilterQueryOk returns a tuple with the FilterQuery field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ParamsPropertyInfraMetricThreshold) GetFilterQueryOk() (*string, bool) {
	if o == nil || IsNil(o.FilterQuery) {
		return nil, false
	}
	return o.FilterQuery, true
}

// HasFilterQuery returns a boolean if a field has been set.
func (o *ParamsPropertyInfraMetricThreshold) HasFilterQuery() bool {
	if o != nil && !IsNil(o.FilterQuery) {
		return true
	}

	return false
}

// SetFilterQuery gets a reference to the given string and assigns it to the FilterQuery field.
func (o *ParamsPropertyInfraMetricThreshold) SetFilterQuery(v string) {
	o.FilterQuery = &v
}

// GetSourceId returns the SourceId field value if set, zero value otherwise.
func (o *ParamsPropertyInfraMetricThreshold) GetSourceId() string {
	if o == nil || IsNil(o.SourceId) {
		var ret string
		return ret
	}
	return *o.SourceId
}

// GetSourceIdOk returns a tuple with the SourceId field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ParamsPropertyInfraMetricThreshold) GetSourceIdOk() (*string, bool) {
	if o == nil || IsNil(o.SourceId) {
		return nil, false
	}
	return o.SourceId, true
}

// HasSourceId returns a boolean if a field has been set.
func (o *ParamsPropertyInfraMetricThreshold) HasSourceId() bool {
	if o != nil && !IsNil(o.SourceId) {
		return true
	}

	return false
}

// SetSourceId gets a reference to the given string and assigns it to the SourceId field.
func (o *ParamsPropertyInfraMetricThreshold) SetSourceId(v string) {
	o.SourceId = &v
}

// GetAlertOnNoData returns the AlertOnNoData field value if set, zero value otherwise.
func (o *ParamsPropertyInfraMetricThreshold) GetAlertOnNoData() bool {
	if o == nil || IsNil(o.AlertOnNoData) {
		var ret bool
		return ret
	}
	return *o.AlertOnNoData
}

// GetAlertOnNoDataOk returns a tuple with the AlertOnNoData field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ParamsPropertyInfraMetricThreshold) GetAlertOnNoDataOk() (*bool, bool) {
	if o == nil || IsNil(o.AlertOnNoData) {
		return nil, false
	}
	return o.AlertOnNoData, true
}

// HasAlertOnNoData returns a boolean if a field has been set.
func (o *ParamsPropertyInfraMetricThreshold) HasAlertOnNoData() bool {
	if o != nil && !IsNil(o.AlertOnNoData) {
		return true
	}

	return false
}

// SetAlertOnNoData gets a reference to the given bool and assigns it to the AlertOnNoData field.
func (o *ParamsPropertyInfraMetricThreshold) SetAlertOnNoData(v bool) {
	o.AlertOnNoData = &v
}

// GetAlertOnGroupDisappear returns the AlertOnGroupDisappear field value if set, zero value otherwise.
func (o *ParamsPropertyInfraMetricThreshold) GetAlertOnGroupDisappear() bool {
	if o == nil || IsNil(o.AlertOnGroupDisappear) {
		var ret bool
		return ret
	}
	return *o.AlertOnGroupDisappear
}

// GetAlertOnGroupDisappearOk returns a tuple with the AlertOnGroupDisappear field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ParamsPropertyInfraMetricThreshold) GetAlertOnGroupDisappearOk() (*bool, bool) {
	if o == nil || IsNil(o.AlertOnGroupDisappear) {
		return nil, false
	}
	return o.AlertOnGroupDisappear, true
}

// HasAlertOnGroupDisappear returns a boolean if a field has been set.
func (o *ParamsPropertyInfraMetricThreshold) HasAlertOnGroupDisappear() bool {
	if o != nil && !IsNil(o.AlertOnGroupDisappear) {
		return true
	}

	return false
}

// SetAlertOnGroupDisappear gets a reference to the given bool and assigns it to the AlertOnGroupDisappear field.
func (o *ParamsPropertyInfraMetricThreshold) SetAlertOnGroupDisappear(v bool) {
	o.AlertOnGroupDisappear = &v
}

func (o ParamsPropertyInfraMetricThreshold) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o ParamsPropertyInfraMetricThreshold) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	if !IsNil(o.Criteria) {
		toSerialize["criteria"] = o.Criteria
	}
	if o.GroupBy.IsSet() {
		toSerialize["groupBy"] = o.GroupBy.Get()
	}
	if !IsNil(o.FilterQuery) {
		toSerialize["filterQuery"] = o.FilterQuery
	}
	if !IsNil(o.SourceId) {
		toSerialize["sourceId"] = o.SourceId
	}
	if !IsNil(o.AlertOnNoData) {
		toSerialize["alertOnNoData"] = o.AlertOnNoData
	}
	if !IsNil(o.AlertOnGroupDisappear) {
		toSerialize["alertOnGroupDisappear"] = o.AlertOnGroupDisappear
	}
	return toSerialize, nil
}

type NullableParamsPropertyInfraMetricThreshold struct {
	value *ParamsPropertyInfraMetricThreshold
	isSet bool
}

func (v NullableParamsPropertyInfraMetricThreshold) Get() *ParamsPropertyInfraMetricThreshold {
	return v.value
}

func (v *NullableParamsPropertyInfraMetricThreshold) Set(val *ParamsPropertyInfraMetricThreshold) {
	v.value = val
	v.isSet = true
}

func (v NullableParamsPropertyInfraMetricThreshold) IsSet() bool {
	return v.isSet
}

func (v *NullableParamsPropertyInfraMetricThreshold) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableParamsPropertyInfraMetricThreshold(val *ParamsPropertyInfraMetricThreshold) *NullableParamsPropertyInfraMetricThreshold {
	return &NullableParamsPropertyInfraMetricThreshold{value: val, isSet: true}
}

func (v NullableParamsPropertyInfraMetricThreshold) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableParamsPropertyInfraMetricThreshold) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
