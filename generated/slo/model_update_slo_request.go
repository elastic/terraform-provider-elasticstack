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

// checks if the UpdateSloRequest type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &UpdateSloRequest{}

// UpdateSloRequest The update SLO API request body varies depending on the type of indicator, time window and budgeting method. Partial update is handled.
type UpdateSloRequest struct {
	// A name for the SLO.
	Name *string `json:"name,omitempty"`
	// A description for the SLO.
	Description     *string                    `json:"description,omitempty"`
	Indicator       *CreateSloRequestIndicator `json:"indicator,omitempty"`
	TimeWindow      *TimeWindow                `json:"timeWindow,omitempty"`
	BudgetingMethod *BudgetingMethod           `json:"budgetingMethod,omitempty"`
	Objective       *Objective                 `json:"objective,omitempty"`
	Settings        *Settings                  `json:"settings,omitempty"`
	GroupBy         *GroupBy                   `json:"groupBy,omitempty"`
	// List of tags
	Tags []string `json:"tags,omitempty"`
}

// NewUpdateSloRequest instantiates a new UpdateSloRequest object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewUpdateSloRequest() *UpdateSloRequest {
	this := UpdateSloRequest{}
	return &this
}

// NewUpdateSloRequestWithDefaults instantiates a new UpdateSloRequest object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewUpdateSloRequestWithDefaults() *UpdateSloRequest {
	this := UpdateSloRequest{}
	return &this
}

// GetName returns the Name field value if set, zero value otherwise.
func (o *UpdateSloRequest) GetName() string {
	if o == nil || IsNil(o.Name) {
		var ret string
		return ret
	}
	return *o.Name
}

// GetNameOk returns a tuple with the Name field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *UpdateSloRequest) GetNameOk() (*string, bool) {
	if o == nil || IsNil(o.Name) {
		return nil, false
	}
	return o.Name, true
}

// HasName returns a boolean if a field has been set.
func (o *UpdateSloRequest) HasName() bool {
	if o != nil && !IsNil(o.Name) {
		return true
	}

	return false
}

// SetName gets a reference to the given string and assigns it to the Name field.
func (o *UpdateSloRequest) SetName(v string) {
	o.Name = &v
}

// GetDescription returns the Description field value if set, zero value otherwise.
func (o *UpdateSloRequest) GetDescription() string {
	if o == nil || IsNil(o.Description) {
		var ret string
		return ret
	}
	return *o.Description
}

// GetDescriptionOk returns a tuple with the Description field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *UpdateSloRequest) GetDescriptionOk() (*string, bool) {
	if o == nil || IsNil(o.Description) {
		return nil, false
	}
	return o.Description, true
}

// HasDescription returns a boolean if a field has been set.
func (o *UpdateSloRequest) HasDescription() bool {
	if o != nil && !IsNil(o.Description) {
		return true
	}

	return false
}

// SetDescription gets a reference to the given string and assigns it to the Description field.
func (o *UpdateSloRequest) SetDescription(v string) {
	o.Description = &v
}

// GetIndicator returns the Indicator field value if set, zero value otherwise.
func (o *UpdateSloRequest) GetIndicator() CreateSloRequestIndicator {
	if o == nil || IsNil(o.Indicator) {
		var ret CreateSloRequestIndicator
		return ret
	}
	return *o.Indicator
}

// GetIndicatorOk returns a tuple with the Indicator field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *UpdateSloRequest) GetIndicatorOk() (*CreateSloRequestIndicator, bool) {
	if o == nil || IsNil(o.Indicator) {
		return nil, false
	}
	return o.Indicator, true
}

// HasIndicator returns a boolean if a field has been set.
func (o *UpdateSloRequest) HasIndicator() bool {
	if o != nil && !IsNil(o.Indicator) {
		return true
	}

	return false
}

// SetIndicator gets a reference to the given CreateSloRequestIndicator and assigns it to the Indicator field.
func (o *UpdateSloRequest) SetIndicator(v CreateSloRequestIndicator) {
	o.Indicator = &v
}

// GetTimeWindow returns the TimeWindow field value if set, zero value otherwise.
func (o *UpdateSloRequest) GetTimeWindow() TimeWindow {
	if o == nil || IsNil(o.TimeWindow) {
		var ret TimeWindow
		return ret
	}
	return *o.TimeWindow
}

// GetTimeWindowOk returns a tuple with the TimeWindow field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *UpdateSloRequest) GetTimeWindowOk() (*TimeWindow, bool) {
	if o == nil || IsNil(o.TimeWindow) {
		return nil, false
	}
	return o.TimeWindow, true
}

// HasTimeWindow returns a boolean if a field has been set.
func (o *UpdateSloRequest) HasTimeWindow() bool {
	if o != nil && !IsNil(o.TimeWindow) {
		return true
	}

	return false
}

// SetTimeWindow gets a reference to the given TimeWindow and assigns it to the TimeWindow field.
func (o *UpdateSloRequest) SetTimeWindow(v TimeWindow) {
	o.TimeWindow = &v
}

// GetBudgetingMethod returns the BudgetingMethod field value if set, zero value otherwise.
func (o *UpdateSloRequest) GetBudgetingMethod() BudgetingMethod {
	if o == nil || IsNil(o.BudgetingMethod) {
		var ret BudgetingMethod
		return ret
	}
	return *o.BudgetingMethod
}

// GetBudgetingMethodOk returns a tuple with the BudgetingMethod field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *UpdateSloRequest) GetBudgetingMethodOk() (*BudgetingMethod, bool) {
	if o == nil || IsNil(o.BudgetingMethod) {
		return nil, false
	}
	return o.BudgetingMethod, true
}

// HasBudgetingMethod returns a boolean if a field has been set.
func (o *UpdateSloRequest) HasBudgetingMethod() bool {
	if o != nil && !IsNil(o.BudgetingMethod) {
		return true
	}

	return false
}

// SetBudgetingMethod gets a reference to the given BudgetingMethod and assigns it to the BudgetingMethod field.
func (o *UpdateSloRequest) SetBudgetingMethod(v BudgetingMethod) {
	o.BudgetingMethod = &v
}

// GetObjective returns the Objective field value if set, zero value otherwise.
func (o *UpdateSloRequest) GetObjective() Objective {
	if o == nil || IsNil(o.Objective) {
		var ret Objective
		return ret
	}
	return *o.Objective
}

// GetObjectiveOk returns a tuple with the Objective field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *UpdateSloRequest) GetObjectiveOk() (*Objective, bool) {
	if o == nil || IsNil(o.Objective) {
		return nil, false
	}
	return o.Objective, true
}

// HasObjective returns a boolean if a field has been set.
func (o *UpdateSloRequest) HasObjective() bool {
	if o != nil && !IsNil(o.Objective) {
		return true
	}

	return false
}

// SetObjective gets a reference to the given Objective and assigns it to the Objective field.
func (o *UpdateSloRequest) SetObjective(v Objective) {
	o.Objective = &v
}

// GetSettings returns the Settings field value if set, zero value otherwise.
func (o *UpdateSloRequest) GetSettings() Settings {
	if o == nil || IsNil(o.Settings) {
		var ret Settings
		return ret
	}
	return *o.Settings
}

// GetSettingsOk returns a tuple with the Settings field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *UpdateSloRequest) GetSettingsOk() (*Settings, bool) {
	if o == nil || IsNil(o.Settings) {
		return nil, false
	}
	return o.Settings, true
}

// HasSettings returns a boolean if a field has been set.
func (o *UpdateSloRequest) HasSettings() bool {
	if o != nil && !IsNil(o.Settings) {
		return true
	}

	return false
}

// SetSettings gets a reference to the given Settings and assigns it to the Settings field.
func (o *UpdateSloRequest) SetSettings(v Settings) {
	o.Settings = &v
}

// GetGroupBy returns the GroupBy field value if set, zero value otherwise.
func (o *UpdateSloRequest) GetGroupBy() GroupBy {
	if o == nil || IsNil(o.GroupBy) {
		var ret GroupBy
		return ret
	}
	return *o.GroupBy
}

// GetGroupByOk returns a tuple with the GroupBy field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *UpdateSloRequest) GetGroupByOk() (*GroupBy, bool) {
	if o == nil || IsNil(o.GroupBy) {
		return nil, false
	}
	return o.GroupBy, true
}

// HasGroupBy returns a boolean if a field has been set.
func (o *UpdateSloRequest) HasGroupBy() bool {
	if o != nil && !IsNil(o.GroupBy) {
		return true
	}

	return false
}

// SetGroupBy gets a reference to the given GroupBy and assigns it to the GroupBy field.
func (o *UpdateSloRequest) SetGroupBy(v GroupBy) {
	o.GroupBy = &v
}

// GetTags returns the Tags field value if set, zero value otherwise.
func (o *UpdateSloRequest) GetTags() []string {
	if o == nil || IsNil(o.Tags) {
		var ret []string
		return ret
	}
	return o.Tags
}

// GetTagsOk returns a tuple with the Tags field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *UpdateSloRequest) GetTagsOk() ([]string, bool) {
	if o == nil || IsNil(o.Tags) {
		return nil, false
	}
	return o.Tags, true
}

// HasTags returns a boolean if a field has been set.
func (o *UpdateSloRequest) HasTags() bool {
	if o != nil && !IsNil(o.Tags) {
		return true
	}

	return false
}

// SetTags gets a reference to the given []string and assigns it to the Tags field.
func (o *UpdateSloRequest) SetTags(v []string) {
	o.Tags = v
}

func (o UpdateSloRequest) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o UpdateSloRequest) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	if !IsNil(o.Name) {
		toSerialize["name"] = o.Name
	}
	if !IsNil(o.Description) {
		toSerialize["description"] = o.Description
	}
	if !IsNil(o.Indicator) {
		toSerialize["indicator"] = o.Indicator
	}
	if !IsNil(o.TimeWindow) {
		toSerialize["timeWindow"] = o.TimeWindow
	}
	if !IsNil(o.BudgetingMethod) {
		toSerialize["budgetingMethod"] = o.BudgetingMethod
	}
	if !IsNil(o.Objective) {
		toSerialize["objective"] = o.Objective
	}
	if !IsNil(o.Settings) {
		toSerialize["settings"] = o.Settings
	}
	if !IsNil(o.GroupBy) {
		toSerialize["groupBy"] = o.GroupBy
	}
	if !IsNil(o.Tags) {
		toSerialize["tags"] = o.Tags
	}
	return toSerialize, nil
}

type NullableUpdateSloRequest struct {
	value *UpdateSloRequest
	isSet bool
}

func (v NullableUpdateSloRequest) Get() *UpdateSloRequest {
	return v.value
}

func (v *NullableUpdateSloRequest) Set(val *UpdateSloRequest) {
	v.value = val
	v.isSet = true
}

func (v NullableUpdateSloRequest) IsSet() bool {
	return v.isSet
}

func (v *NullableUpdateSloRequest) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableUpdateSloRequest(val *UpdateSloRequest) *NullableUpdateSloRequest {
	return &NullableUpdateSloRequest{value: val, isSet: true}
}

func (v NullableUpdateSloRequest) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableUpdateSloRequest) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
