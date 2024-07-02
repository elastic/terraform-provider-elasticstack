/*
Alerting

OpenAPI schema for alerting endpoints

API version: 0.2
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package alerting

import (
	"encoding/json"
	"time"
)

// checks if the AlertResponseProperties type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &AlertResponseProperties{}

// AlertResponseProperties struct for AlertResponseProperties
type AlertResponseProperties struct {
	Actions     []map[string]interface{} `json:"actions,omitempty"`
	AlertTypeId *string                  `json:"alertTypeId,omitempty"`
	ApiKeyOwner interface{}              `json:"apiKeyOwner,omitempty"`
	// The date and time that the alert was created.
	CreatedAt *time.Time `json:"createdAt,omitempty"`
	// The identifier for the user that created the alert.
	CreatedBy *string `json:"createdBy,omitempty"`
	// Indicates whether the alert is currently enabled.
	Enabled         *bool                                   `json:"enabled,omitempty"`
	ExecutionStatus *AlertResponsePropertiesExecutionStatus `json:"executionStatus,omitempty"`
	// The identifier for the alert.
	Id               *string  `json:"id,omitempty"`
	MuteAll          *bool    `json:"muteAll,omitempty"`
	MutedInstanceIds []string `json:"mutedInstanceIds,omitempty"`
	// The name of the alert.
	Name            *string                          `json:"name,omitempty"`
	NotifyWhen      *string                          `json:"notifyWhen,omitempty"`
	Params          map[string]interface{}           `json:"params,omitempty"`
	Schedule        *AlertResponsePropertiesSchedule `json:"schedule,omitempty"`
	ScheduledTaskId *string                          `json:"scheduledTaskId,omitempty"`
	Tags            []string                         `json:"tags,omitempty"`
	Throttle        interface{}                      `json:"throttle,omitempty"`
	UpdatedAt       *string                          `json:"updatedAt,omitempty"`
	// The identifier for the user that updated this alert most recently.
	UpdatedBy interface{} `json:"updatedBy,omitempty"`
}

// NewAlertResponseProperties instantiates a new AlertResponseProperties object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewAlertResponseProperties() *AlertResponseProperties {
	this := AlertResponseProperties{}
	return &this
}

// NewAlertResponsePropertiesWithDefaults instantiates a new AlertResponseProperties object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewAlertResponsePropertiesWithDefaults() *AlertResponseProperties {
	this := AlertResponseProperties{}
	return &this
}

// GetActions returns the Actions field value if set, zero value otherwise.
func (o *AlertResponseProperties) GetActions() []map[string]interface{} {
	if o == nil || IsNil(o.Actions) {
		var ret []map[string]interface{}
		return ret
	}
	return o.Actions
}

// GetActionsOk returns a tuple with the Actions field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *AlertResponseProperties) GetActionsOk() ([]map[string]interface{}, bool) {
	if o == nil || IsNil(o.Actions) {
		return nil, false
	}
	return o.Actions, true
}

// HasActions returns a boolean if a field has been set.
func (o *AlertResponseProperties) HasActions() bool {
	if o != nil && !IsNil(o.Actions) {
		return true
	}

	return false
}

// SetActions gets a reference to the given []map[string]interface{} and assigns it to the Actions field.
func (o *AlertResponseProperties) SetActions(v []map[string]interface{}) {
	o.Actions = v
}

// GetAlertTypeId returns the AlertTypeId field value if set, zero value otherwise.
func (o *AlertResponseProperties) GetAlertTypeId() string {
	if o == nil || IsNil(o.AlertTypeId) {
		var ret string
		return ret
	}
	return *o.AlertTypeId
}

// GetAlertTypeIdOk returns a tuple with the AlertTypeId field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *AlertResponseProperties) GetAlertTypeIdOk() (*string, bool) {
	if o == nil || IsNil(o.AlertTypeId) {
		return nil, false
	}
	return o.AlertTypeId, true
}

// HasAlertTypeId returns a boolean if a field has been set.
func (o *AlertResponseProperties) HasAlertTypeId() bool {
	if o != nil && !IsNil(o.AlertTypeId) {
		return true
	}

	return false
}

// SetAlertTypeId gets a reference to the given string and assigns it to the AlertTypeId field.
func (o *AlertResponseProperties) SetAlertTypeId(v string) {
	o.AlertTypeId = &v
}

// GetApiKeyOwner returns the ApiKeyOwner field value if set, zero value otherwise (both if not set or set to explicit null).
func (o *AlertResponseProperties) GetApiKeyOwner() interface{} {
	if o == nil {
		var ret interface{}
		return ret
	}
	return o.ApiKeyOwner
}

// GetApiKeyOwnerOk returns a tuple with the ApiKeyOwner field value if set, nil otherwise
// and a boolean to check if the value has been set.
// NOTE: If the value is an explicit nil, `nil, true` will be returned
func (o *AlertResponseProperties) GetApiKeyOwnerOk() (*interface{}, bool) {
	if o == nil || IsNil(o.ApiKeyOwner) {
		return nil, false
	}
	return &o.ApiKeyOwner, true
}

// HasApiKeyOwner returns a boolean if a field has been set.
func (o *AlertResponseProperties) HasApiKeyOwner() bool {
	if o != nil && IsNil(o.ApiKeyOwner) {
		return true
	}

	return false
}

// SetApiKeyOwner gets a reference to the given interface{} and assigns it to the ApiKeyOwner field.
func (o *AlertResponseProperties) SetApiKeyOwner(v interface{}) {
	o.ApiKeyOwner = v
}

// GetCreatedAt returns the CreatedAt field value if set, zero value otherwise.
func (o *AlertResponseProperties) GetCreatedAt() time.Time {
	if o == nil || IsNil(o.CreatedAt) {
		var ret time.Time
		return ret
	}
	return *o.CreatedAt
}

// GetCreatedAtOk returns a tuple with the CreatedAt field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *AlertResponseProperties) GetCreatedAtOk() (*time.Time, bool) {
	if o == nil || IsNil(o.CreatedAt) {
		return nil, false
	}
	return o.CreatedAt, true
}

// HasCreatedAt returns a boolean if a field has been set.
func (o *AlertResponseProperties) HasCreatedAt() bool {
	if o != nil && !IsNil(o.CreatedAt) {
		return true
	}

	return false
}

// SetCreatedAt gets a reference to the given time.Time and assigns it to the CreatedAt field.
func (o *AlertResponseProperties) SetCreatedAt(v time.Time) {
	o.CreatedAt = &v
}

// GetCreatedBy returns the CreatedBy field value if set, zero value otherwise.
func (o *AlertResponseProperties) GetCreatedBy() string {
	if o == nil || IsNil(o.CreatedBy) {
		var ret string
		return ret
	}
	return *o.CreatedBy
}

// GetCreatedByOk returns a tuple with the CreatedBy field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *AlertResponseProperties) GetCreatedByOk() (*string, bool) {
	if o == nil || IsNil(o.CreatedBy) {
		return nil, false
	}
	return o.CreatedBy, true
}

// HasCreatedBy returns a boolean if a field has been set.
func (o *AlertResponseProperties) HasCreatedBy() bool {
	if o != nil && !IsNil(o.CreatedBy) {
		return true
	}

	return false
}

// SetCreatedBy gets a reference to the given string and assigns it to the CreatedBy field.
func (o *AlertResponseProperties) SetCreatedBy(v string) {
	o.CreatedBy = &v
}

// GetEnabled returns the Enabled field value if set, zero value otherwise.
func (o *AlertResponseProperties) GetEnabled() bool {
	if o == nil || IsNil(o.Enabled) {
		var ret bool
		return ret
	}
	return *o.Enabled
}

// GetEnabledOk returns a tuple with the Enabled field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *AlertResponseProperties) GetEnabledOk() (*bool, bool) {
	if o == nil || IsNil(o.Enabled) {
		return nil, false
	}
	return o.Enabled, true
}

// HasEnabled returns a boolean if a field has been set.
func (o *AlertResponseProperties) HasEnabled() bool {
	if o != nil && !IsNil(o.Enabled) {
		return true
	}

	return false
}

// SetEnabled gets a reference to the given bool and assigns it to the Enabled field.
func (o *AlertResponseProperties) SetEnabled(v bool) {
	o.Enabled = &v
}

// GetExecutionStatus returns the ExecutionStatus field value if set, zero value otherwise.
func (o *AlertResponseProperties) GetExecutionStatus() AlertResponsePropertiesExecutionStatus {
	if o == nil || IsNil(o.ExecutionStatus) {
		var ret AlertResponsePropertiesExecutionStatus
		return ret
	}
	return *o.ExecutionStatus
}

// GetExecutionStatusOk returns a tuple with the ExecutionStatus field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *AlertResponseProperties) GetExecutionStatusOk() (*AlertResponsePropertiesExecutionStatus, bool) {
	if o == nil || IsNil(o.ExecutionStatus) {
		return nil, false
	}
	return o.ExecutionStatus, true
}

// HasExecutionStatus returns a boolean if a field has been set.
func (o *AlertResponseProperties) HasExecutionStatus() bool {
	if o != nil && !IsNil(o.ExecutionStatus) {
		return true
	}

	return false
}

// SetExecutionStatus gets a reference to the given AlertResponsePropertiesExecutionStatus and assigns it to the ExecutionStatus field.
func (o *AlertResponseProperties) SetExecutionStatus(v AlertResponsePropertiesExecutionStatus) {
	o.ExecutionStatus = &v
}

// GetId returns the Id field value if set, zero value otherwise.
func (o *AlertResponseProperties) GetId() string {
	if o == nil || IsNil(o.Id) {
		var ret string
		return ret
	}
	return *o.Id
}

// GetIdOk returns a tuple with the Id field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *AlertResponseProperties) GetIdOk() (*string, bool) {
	if o == nil || IsNil(o.Id) {
		return nil, false
	}
	return o.Id, true
}

// HasId returns a boolean if a field has been set.
func (o *AlertResponseProperties) HasId() bool {
	if o != nil && !IsNil(o.Id) {
		return true
	}

	return false
}

// SetId gets a reference to the given string and assigns it to the Id field.
func (o *AlertResponseProperties) SetId(v string) {
	o.Id = &v
}

// GetMuteAll returns the MuteAll field value if set, zero value otherwise.
func (o *AlertResponseProperties) GetMuteAll() bool {
	if o == nil || IsNil(o.MuteAll) {
		var ret bool
		return ret
	}
	return *o.MuteAll
}

// GetMuteAllOk returns a tuple with the MuteAll field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *AlertResponseProperties) GetMuteAllOk() (*bool, bool) {
	if o == nil || IsNil(o.MuteAll) {
		return nil, false
	}
	return o.MuteAll, true
}

// HasMuteAll returns a boolean if a field has been set.
func (o *AlertResponseProperties) HasMuteAll() bool {
	if o != nil && !IsNil(o.MuteAll) {
		return true
	}

	return false
}

// SetMuteAll gets a reference to the given bool and assigns it to the MuteAll field.
func (o *AlertResponseProperties) SetMuteAll(v bool) {
	o.MuteAll = &v
}

// GetMutedInstanceIds returns the MutedInstanceIds field value if set, zero value otherwise.
func (o *AlertResponseProperties) GetMutedInstanceIds() []string {
	if o == nil || IsNil(o.MutedInstanceIds) {
		var ret []string
		return ret
	}
	return o.MutedInstanceIds
}

// GetMutedInstanceIdsOk returns a tuple with the MutedInstanceIds field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *AlertResponseProperties) GetMutedInstanceIdsOk() ([]string, bool) {
	if o == nil || IsNil(o.MutedInstanceIds) {
		return nil, false
	}
	return o.MutedInstanceIds, true
}

// HasMutedInstanceIds returns a boolean if a field has been set.
func (o *AlertResponseProperties) HasMutedInstanceIds() bool {
	if o != nil && !IsNil(o.MutedInstanceIds) {
		return true
	}

	return false
}

// SetMutedInstanceIds gets a reference to the given []string and assigns it to the MutedInstanceIds field.
func (o *AlertResponseProperties) SetMutedInstanceIds(v []string) {
	o.MutedInstanceIds = v
}

// GetName returns the Name field value if set, zero value otherwise.
func (o *AlertResponseProperties) GetName() string {
	if o == nil || IsNil(o.Name) {
		var ret string
		return ret
	}
	return *o.Name
}

// GetNameOk returns a tuple with the Name field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *AlertResponseProperties) GetNameOk() (*string, bool) {
	if o == nil || IsNil(o.Name) {
		return nil, false
	}
	return o.Name, true
}

// HasName returns a boolean if a field has been set.
func (o *AlertResponseProperties) HasName() bool {
	if o != nil && !IsNil(o.Name) {
		return true
	}

	return false
}

// SetName gets a reference to the given string and assigns it to the Name field.
func (o *AlertResponseProperties) SetName(v string) {
	o.Name = &v
}

// GetNotifyWhen returns the NotifyWhen field value if set, zero value otherwise.
func (o *AlertResponseProperties) GetNotifyWhen() string {
	if o == nil || IsNil(o.NotifyWhen) {
		var ret string
		return ret
	}
	return *o.NotifyWhen
}

// GetNotifyWhenOk returns a tuple with the NotifyWhen field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *AlertResponseProperties) GetNotifyWhenOk() (*string, bool) {
	if o == nil || IsNil(o.NotifyWhen) {
		return nil, false
	}
	return o.NotifyWhen, true
}

// HasNotifyWhen returns a boolean if a field has been set.
func (o *AlertResponseProperties) HasNotifyWhen() bool {
	if o != nil && !IsNil(o.NotifyWhen) {
		return true
	}

	return false
}

// SetNotifyWhen gets a reference to the given string and assigns it to the NotifyWhen field.
func (o *AlertResponseProperties) SetNotifyWhen(v string) {
	o.NotifyWhen = &v
}

// GetParams returns the Params field value if set, zero value otherwise.
func (o *AlertResponseProperties) GetParams() map[string]interface{} {
	if o == nil || IsNil(o.Params) {
		var ret map[string]interface{}
		return ret
	}
	return o.Params
}

// GetParamsOk returns a tuple with the Params field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *AlertResponseProperties) GetParamsOk() (map[string]interface{}, bool) {
	if o == nil || IsNil(o.Params) {
		return map[string]interface{}{}, false
	}
	return o.Params, true
}

// HasParams returns a boolean if a field has been set.
func (o *AlertResponseProperties) HasParams() bool {
	if o != nil && !IsNil(o.Params) {
		return true
	}

	return false
}

// SetParams gets a reference to the given map[string]interface{} and assigns it to the Params field.
func (o *AlertResponseProperties) SetParams(v map[string]interface{}) {
	o.Params = v
}

// GetSchedule returns the Schedule field value if set, zero value otherwise.
func (o *AlertResponseProperties) GetSchedule() AlertResponsePropertiesSchedule {
	if o == nil || IsNil(o.Schedule) {
		var ret AlertResponsePropertiesSchedule
		return ret
	}
	return *o.Schedule
}

// GetScheduleOk returns a tuple with the Schedule field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *AlertResponseProperties) GetScheduleOk() (*AlertResponsePropertiesSchedule, bool) {
	if o == nil || IsNil(o.Schedule) {
		return nil, false
	}
	return o.Schedule, true
}

// HasSchedule returns a boolean if a field has been set.
func (o *AlertResponseProperties) HasSchedule() bool {
	if o != nil && !IsNil(o.Schedule) {
		return true
	}

	return false
}

// SetSchedule gets a reference to the given AlertResponsePropertiesSchedule and assigns it to the Schedule field.
func (o *AlertResponseProperties) SetSchedule(v AlertResponsePropertiesSchedule) {
	o.Schedule = &v
}

// GetScheduledTaskId returns the ScheduledTaskId field value if set, zero value otherwise.
func (o *AlertResponseProperties) GetScheduledTaskId() string {
	if o == nil || IsNil(o.ScheduledTaskId) {
		var ret string
		return ret
	}
	return *o.ScheduledTaskId
}

// GetScheduledTaskIdOk returns a tuple with the ScheduledTaskId field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *AlertResponseProperties) GetScheduledTaskIdOk() (*string, bool) {
	if o == nil || IsNil(o.ScheduledTaskId) {
		return nil, false
	}
	return o.ScheduledTaskId, true
}

// HasScheduledTaskId returns a boolean if a field has been set.
func (o *AlertResponseProperties) HasScheduledTaskId() bool {
	if o != nil && !IsNil(o.ScheduledTaskId) {
		return true
	}

	return false
}

// SetScheduledTaskId gets a reference to the given string and assigns it to the ScheduledTaskId field.
func (o *AlertResponseProperties) SetScheduledTaskId(v string) {
	o.ScheduledTaskId = &v
}

// GetTags returns the Tags field value if set, zero value otherwise.
func (o *AlertResponseProperties) GetTags() []string {
	if o == nil || IsNil(o.Tags) {
		var ret []string
		return ret
	}
	return o.Tags
}

// GetTagsOk returns a tuple with the Tags field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *AlertResponseProperties) GetTagsOk() ([]string, bool) {
	if o == nil || IsNil(o.Tags) {
		return nil, false
	}
	return o.Tags, true
}

// HasTags returns a boolean if a field has been set.
func (o *AlertResponseProperties) HasTags() bool {
	if o != nil && !IsNil(o.Tags) {
		return true
	}

	return false
}

// SetTags gets a reference to the given []string and assigns it to the Tags field.
func (o *AlertResponseProperties) SetTags(v []string) {
	o.Tags = v
}

// GetThrottle returns the Throttle field value if set, zero value otherwise (both if not set or set to explicit null).
func (o *AlertResponseProperties) GetThrottle() interface{} {
	if o == nil {
		var ret interface{}
		return ret
	}
	return o.Throttle
}

// GetThrottleOk returns a tuple with the Throttle field value if set, nil otherwise
// and a boolean to check if the value has been set.
// NOTE: If the value is an explicit nil, `nil, true` will be returned
func (o *AlertResponseProperties) GetThrottleOk() (*interface{}, bool) {
	if o == nil || IsNil(o.Throttle) {
		return nil, false
	}
	return &o.Throttle, true
}

// HasThrottle returns a boolean if a field has been set.
func (o *AlertResponseProperties) HasThrottle() bool {
	if o != nil && IsNil(o.Throttle) {
		return true
	}

	return false
}

// SetThrottle gets a reference to the given interface{} and assigns it to the Throttle field.
func (o *AlertResponseProperties) SetThrottle(v interface{}) {
	o.Throttle = v
}

// GetUpdatedAt returns the UpdatedAt field value if set, zero value otherwise.
func (o *AlertResponseProperties) GetUpdatedAt() string {
	if o == nil || IsNil(o.UpdatedAt) {
		var ret string
		return ret
	}
	return *o.UpdatedAt
}

// GetUpdatedAtOk returns a tuple with the UpdatedAt field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *AlertResponseProperties) GetUpdatedAtOk() (*string, bool) {
	if o == nil || IsNil(o.UpdatedAt) {
		return nil, false
	}
	return o.UpdatedAt, true
}

// HasUpdatedAt returns a boolean if a field has been set.
func (o *AlertResponseProperties) HasUpdatedAt() bool {
	if o != nil && !IsNil(o.UpdatedAt) {
		return true
	}

	return false
}

// SetUpdatedAt gets a reference to the given string and assigns it to the UpdatedAt field.
func (o *AlertResponseProperties) SetUpdatedAt(v string) {
	o.UpdatedAt = &v
}

// GetUpdatedBy returns the UpdatedBy field value if set, zero value otherwise (both if not set or set to explicit null).
func (o *AlertResponseProperties) GetUpdatedBy() interface{} {
	if o == nil {
		var ret interface{}
		return ret
	}
	return o.UpdatedBy
}

// GetUpdatedByOk returns a tuple with the UpdatedBy field value if set, nil otherwise
// and a boolean to check if the value has been set.
// NOTE: If the value is an explicit nil, `nil, true` will be returned
func (o *AlertResponseProperties) GetUpdatedByOk() (*interface{}, bool) {
	if o == nil || IsNil(o.UpdatedBy) {
		return nil, false
	}
	return &o.UpdatedBy, true
}

// HasUpdatedBy returns a boolean if a field has been set.
func (o *AlertResponseProperties) HasUpdatedBy() bool {
	if o != nil && IsNil(o.UpdatedBy) {
		return true
	}

	return false
}

// SetUpdatedBy gets a reference to the given interface{} and assigns it to the UpdatedBy field.
func (o *AlertResponseProperties) SetUpdatedBy(v interface{}) {
	o.UpdatedBy = v
}

func (o AlertResponseProperties) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o AlertResponseProperties) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	if !IsNil(o.Actions) {
		toSerialize["actions"] = o.Actions
	}
	if !IsNil(o.AlertTypeId) {
		toSerialize["alertTypeId"] = o.AlertTypeId
	}
	if o.ApiKeyOwner != nil {
		toSerialize["apiKeyOwner"] = o.ApiKeyOwner
	}
	if !IsNil(o.CreatedAt) {
		toSerialize["createdAt"] = o.CreatedAt
	}
	if !IsNil(o.CreatedBy) {
		toSerialize["createdBy"] = o.CreatedBy
	}
	if !IsNil(o.Enabled) {
		toSerialize["enabled"] = o.Enabled
	}
	if !IsNil(o.ExecutionStatus) {
		toSerialize["executionStatus"] = o.ExecutionStatus
	}
	if !IsNil(o.Id) {
		toSerialize["id"] = o.Id
	}
	if !IsNil(o.MuteAll) {
		toSerialize["muteAll"] = o.MuteAll
	}
	if !IsNil(o.MutedInstanceIds) {
		toSerialize["mutedInstanceIds"] = o.MutedInstanceIds
	}
	if !IsNil(o.Name) {
		toSerialize["name"] = o.Name
	}
	if !IsNil(o.NotifyWhen) {
		toSerialize["notifyWhen"] = o.NotifyWhen
	}
	if !IsNil(o.Params) {
		toSerialize["params"] = o.Params
	}
	if !IsNil(o.Schedule) {
		toSerialize["schedule"] = o.Schedule
	}
	if !IsNil(o.ScheduledTaskId) {
		toSerialize["scheduledTaskId"] = o.ScheduledTaskId
	}
	if !IsNil(o.Tags) {
		toSerialize["tags"] = o.Tags
	}
	if o.Throttle != nil {
		toSerialize["throttle"] = o.Throttle
	}
	if !IsNil(o.UpdatedAt) {
		toSerialize["updatedAt"] = o.UpdatedAt
	}
	if o.UpdatedBy != nil {
		toSerialize["updatedBy"] = o.UpdatedBy
	}
	return toSerialize, nil
}

type NullableAlertResponseProperties struct {
	value *AlertResponseProperties
	isSet bool
}

func (v NullableAlertResponseProperties) Get() *AlertResponseProperties {
	return v.value
}

func (v *NullableAlertResponseProperties) Set(val *AlertResponseProperties) {
	v.value = val
	v.isSet = true
}

func (v NullableAlertResponseProperties) IsSet() bool {
	return v.isSet
}

func (v *NullableAlertResponseProperties) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableAlertResponseProperties(val *AlertResponseProperties) *NullableAlertResponseProperties {
	return &NullableAlertResponseProperties{value: val, isSet: true}
}

func (v NullableAlertResponseProperties) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableAlertResponseProperties) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
