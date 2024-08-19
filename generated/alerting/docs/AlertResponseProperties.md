# AlertResponseProperties

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Actions** | Pointer to **[]map[string]interface{}** |  | [optional] 
**AlertTypeId** | Pointer to **string** |  | [optional] 
**ApiKeyOwner** | Pointer to **NullableString** |  | [optional] 
**CreatedAt** | Pointer to **time.Time** | The date and time that the alert was created. | [optional] 
**CreatedBy** | Pointer to **string** | The identifier for the user that created the alert. | [optional] 
**Enabled** | Pointer to **bool** | Indicates whether the alert is currently enabled. | [optional] 
**ExecutionStatus** | Pointer to [**AlertResponsePropertiesExecutionStatus**](AlertResponsePropertiesExecutionStatus.md) |  | [optional] 
**Id** | Pointer to **string** | The identifier for the alert. | [optional] 
**MuteAll** | Pointer to **bool** |  | [optional] 
**MutedInstanceIds** | Pointer to **[]string** |  | [optional] 
**Name** | Pointer to **string** | The name of the alert. | [optional] 
**NotifyWhen** | Pointer to **string** |  | [optional] 
**Params** | Pointer to **map[string]interface{}** |  | [optional] 
**Schedule** | Pointer to [**AlertResponsePropertiesSchedule**](AlertResponsePropertiesSchedule.md) |  | [optional] 
**ScheduledTaskId** | Pointer to **string** |  | [optional] 
**Tags** | Pointer to **[]string** |  | [optional] 
**Throttle** | Pointer to **NullableString** |  | [optional] 
**UpdatedAt** | Pointer to **string** |  | [optional] 
**UpdatedBy** | Pointer to **NullableString** | The identifier for the user that updated this alert most recently. | [optional] 

## Methods

### NewAlertResponseProperties

`func NewAlertResponseProperties() *AlertResponseProperties`

NewAlertResponseProperties instantiates a new AlertResponseProperties object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAlertResponsePropertiesWithDefaults

`func NewAlertResponsePropertiesWithDefaults() *AlertResponseProperties`

NewAlertResponsePropertiesWithDefaults instantiates a new AlertResponseProperties object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetActions

`func (o *AlertResponseProperties) GetActions() []map[string]interface{}`

GetActions returns the Actions field if non-nil, zero value otherwise.

### GetActionsOk

`func (o *AlertResponseProperties) GetActionsOk() (*[]map[string]interface{}, bool)`

GetActionsOk returns a tuple with the Actions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetActions

`func (o *AlertResponseProperties) SetActions(v []map[string]interface{})`

SetActions sets Actions field to given value.

### HasActions

`func (o *AlertResponseProperties) HasActions() bool`

HasActions returns a boolean if a field has been set.

### GetAlertTypeId

`func (o *AlertResponseProperties) GetAlertTypeId() string`

GetAlertTypeId returns the AlertTypeId field if non-nil, zero value otherwise.

### GetAlertTypeIdOk

`func (o *AlertResponseProperties) GetAlertTypeIdOk() (*string, bool)`

GetAlertTypeIdOk returns a tuple with the AlertTypeId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAlertTypeId

`func (o *AlertResponseProperties) SetAlertTypeId(v string)`

SetAlertTypeId sets AlertTypeId field to given value.

### HasAlertTypeId

`func (o *AlertResponseProperties) HasAlertTypeId() bool`

HasAlertTypeId returns a boolean if a field has been set.

### GetApiKeyOwner

`func (o *AlertResponseProperties) GetApiKeyOwner() string`

GetApiKeyOwner returns the ApiKeyOwner field if non-nil, zero value otherwise.

### GetApiKeyOwnerOk

`func (o *AlertResponseProperties) GetApiKeyOwnerOk() (*string, bool)`

GetApiKeyOwnerOk returns a tuple with the ApiKeyOwner field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetApiKeyOwner

`func (o *AlertResponseProperties) SetApiKeyOwner(v string)`

SetApiKeyOwner sets ApiKeyOwner field to given value.

### HasApiKeyOwner

`func (o *AlertResponseProperties) HasApiKeyOwner() bool`

HasApiKeyOwner returns a boolean if a field has been set.

### SetApiKeyOwnerNil

`func (o *AlertResponseProperties) SetApiKeyOwnerNil(b bool)`

 SetApiKeyOwnerNil sets the value for ApiKeyOwner to be an explicit nil

### UnsetApiKeyOwner
`func (o *AlertResponseProperties) UnsetApiKeyOwner()`

UnsetApiKeyOwner ensures that no value is present for ApiKeyOwner, not even an explicit nil
### GetCreatedAt

`func (o *AlertResponseProperties) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *AlertResponseProperties) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *AlertResponseProperties) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.

### HasCreatedAt

`func (o *AlertResponseProperties) HasCreatedAt() bool`

HasCreatedAt returns a boolean if a field has been set.

### GetCreatedBy

`func (o *AlertResponseProperties) GetCreatedBy() string`

GetCreatedBy returns the CreatedBy field if non-nil, zero value otherwise.

### GetCreatedByOk

`func (o *AlertResponseProperties) GetCreatedByOk() (*string, bool)`

GetCreatedByOk returns a tuple with the CreatedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedBy

`func (o *AlertResponseProperties) SetCreatedBy(v string)`

SetCreatedBy sets CreatedBy field to given value.

### HasCreatedBy

`func (o *AlertResponseProperties) HasCreatedBy() bool`

HasCreatedBy returns a boolean if a field has been set.

### GetEnabled

`func (o *AlertResponseProperties) GetEnabled() bool`

GetEnabled returns the Enabled field if non-nil, zero value otherwise.

### GetEnabledOk

`func (o *AlertResponseProperties) GetEnabledOk() (*bool, bool)`

GetEnabledOk returns a tuple with the Enabled field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnabled

`func (o *AlertResponseProperties) SetEnabled(v bool)`

SetEnabled sets Enabled field to given value.

### HasEnabled

`func (o *AlertResponseProperties) HasEnabled() bool`

HasEnabled returns a boolean if a field has been set.

### GetExecutionStatus

`func (o *AlertResponseProperties) GetExecutionStatus() AlertResponsePropertiesExecutionStatus`

GetExecutionStatus returns the ExecutionStatus field if non-nil, zero value otherwise.

### GetExecutionStatusOk

`func (o *AlertResponseProperties) GetExecutionStatusOk() (*AlertResponsePropertiesExecutionStatus, bool)`

GetExecutionStatusOk returns a tuple with the ExecutionStatus field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExecutionStatus

`func (o *AlertResponseProperties) SetExecutionStatus(v AlertResponsePropertiesExecutionStatus)`

SetExecutionStatus sets ExecutionStatus field to given value.

### HasExecutionStatus

`func (o *AlertResponseProperties) HasExecutionStatus() bool`

HasExecutionStatus returns a boolean if a field has been set.

### GetId

`func (o *AlertResponseProperties) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *AlertResponseProperties) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *AlertResponseProperties) SetId(v string)`

SetId sets Id field to given value.

### HasId

`func (o *AlertResponseProperties) HasId() bool`

HasId returns a boolean if a field has been set.

### GetMuteAll

`func (o *AlertResponseProperties) GetMuteAll() bool`

GetMuteAll returns the MuteAll field if non-nil, zero value otherwise.

### GetMuteAllOk

`func (o *AlertResponseProperties) GetMuteAllOk() (*bool, bool)`

GetMuteAllOk returns a tuple with the MuteAll field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMuteAll

`func (o *AlertResponseProperties) SetMuteAll(v bool)`

SetMuteAll sets MuteAll field to given value.

### HasMuteAll

`func (o *AlertResponseProperties) HasMuteAll() bool`

HasMuteAll returns a boolean if a field has been set.

### GetMutedInstanceIds

`func (o *AlertResponseProperties) GetMutedInstanceIds() []string`

GetMutedInstanceIds returns the MutedInstanceIds field if non-nil, zero value otherwise.

### GetMutedInstanceIdsOk

`func (o *AlertResponseProperties) GetMutedInstanceIdsOk() (*[]string, bool)`

GetMutedInstanceIdsOk returns a tuple with the MutedInstanceIds field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMutedInstanceIds

`func (o *AlertResponseProperties) SetMutedInstanceIds(v []string)`

SetMutedInstanceIds sets MutedInstanceIds field to given value.

### HasMutedInstanceIds

`func (o *AlertResponseProperties) HasMutedInstanceIds() bool`

HasMutedInstanceIds returns a boolean if a field has been set.

### GetName

`func (o *AlertResponseProperties) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *AlertResponseProperties) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *AlertResponseProperties) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *AlertResponseProperties) HasName() bool`

HasName returns a boolean if a field has been set.

### GetNotifyWhen

`func (o *AlertResponseProperties) GetNotifyWhen() string`

GetNotifyWhen returns the NotifyWhen field if non-nil, zero value otherwise.

### GetNotifyWhenOk

`func (o *AlertResponseProperties) GetNotifyWhenOk() (*string, bool)`

GetNotifyWhenOk returns a tuple with the NotifyWhen field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNotifyWhen

`func (o *AlertResponseProperties) SetNotifyWhen(v string)`

SetNotifyWhen sets NotifyWhen field to given value.

### HasNotifyWhen

`func (o *AlertResponseProperties) HasNotifyWhen() bool`

HasNotifyWhen returns a boolean if a field has been set.

### GetParams

`func (o *AlertResponseProperties) GetParams() map[string]interface{}`

GetParams returns the Params field if non-nil, zero value otherwise.

### GetParamsOk

`func (o *AlertResponseProperties) GetParamsOk() (*map[string]interface{}, bool)`

GetParamsOk returns a tuple with the Params field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetParams

`func (o *AlertResponseProperties) SetParams(v map[string]interface{})`

SetParams sets Params field to given value.

### HasParams

`func (o *AlertResponseProperties) HasParams() bool`

HasParams returns a boolean if a field has been set.

### GetSchedule

`func (o *AlertResponseProperties) GetSchedule() AlertResponsePropertiesSchedule`

GetSchedule returns the Schedule field if non-nil, zero value otherwise.

### GetScheduleOk

`func (o *AlertResponseProperties) GetScheduleOk() (*AlertResponsePropertiesSchedule, bool)`

GetScheduleOk returns a tuple with the Schedule field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSchedule

`func (o *AlertResponseProperties) SetSchedule(v AlertResponsePropertiesSchedule)`

SetSchedule sets Schedule field to given value.

### HasSchedule

`func (o *AlertResponseProperties) HasSchedule() bool`

HasSchedule returns a boolean if a field has been set.

### GetScheduledTaskId

`func (o *AlertResponseProperties) GetScheduledTaskId() string`

GetScheduledTaskId returns the ScheduledTaskId field if non-nil, zero value otherwise.

### GetScheduledTaskIdOk

`func (o *AlertResponseProperties) GetScheduledTaskIdOk() (*string, bool)`

GetScheduledTaskIdOk returns a tuple with the ScheduledTaskId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScheduledTaskId

`func (o *AlertResponseProperties) SetScheduledTaskId(v string)`

SetScheduledTaskId sets ScheduledTaskId field to given value.

### HasScheduledTaskId

`func (o *AlertResponseProperties) HasScheduledTaskId() bool`

HasScheduledTaskId returns a boolean if a field has been set.

### GetTags

`func (o *AlertResponseProperties) GetTags() []string`

GetTags returns the Tags field if non-nil, zero value otherwise.

### GetTagsOk

`func (o *AlertResponseProperties) GetTagsOk() (*[]string, bool)`

GetTagsOk returns a tuple with the Tags field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTags

`func (o *AlertResponseProperties) SetTags(v []string)`

SetTags sets Tags field to given value.

### HasTags

`func (o *AlertResponseProperties) HasTags() bool`

HasTags returns a boolean if a field has been set.

### GetThrottle

`func (o *AlertResponseProperties) GetThrottle() string`

GetThrottle returns the Throttle field if non-nil, zero value otherwise.

### GetThrottleOk

`func (o *AlertResponseProperties) GetThrottleOk() (*string, bool)`

GetThrottleOk returns a tuple with the Throttle field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetThrottle

`func (o *AlertResponseProperties) SetThrottle(v string)`

SetThrottle sets Throttle field to given value.

### HasThrottle

`func (o *AlertResponseProperties) HasThrottle() bool`

HasThrottle returns a boolean if a field has been set.

### SetThrottleNil

`func (o *AlertResponseProperties) SetThrottleNil(b bool)`

 SetThrottleNil sets the value for Throttle to be an explicit nil

### UnsetThrottle
`func (o *AlertResponseProperties) UnsetThrottle()`

UnsetThrottle ensures that no value is present for Throttle, not even an explicit nil
### GetUpdatedAt

`func (o *AlertResponseProperties) GetUpdatedAt() string`

GetUpdatedAt returns the UpdatedAt field if non-nil, zero value otherwise.

### GetUpdatedAtOk

`func (o *AlertResponseProperties) GetUpdatedAtOk() (*string, bool)`

GetUpdatedAtOk returns a tuple with the UpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedAt

`func (o *AlertResponseProperties) SetUpdatedAt(v string)`

SetUpdatedAt sets UpdatedAt field to given value.

### HasUpdatedAt

`func (o *AlertResponseProperties) HasUpdatedAt() bool`

HasUpdatedAt returns a boolean if a field has been set.

### GetUpdatedBy

`func (o *AlertResponseProperties) GetUpdatedBy() string`

GetUpdatedBy returns the UpdatedBy field if non-nil, zero value otherwise.

### GetUpdatedByOk

`func (o *AlertResponseProperties) GetUpdatedByOk() (*string, bool)`

GetUpdatedByOk returns a tuple with the UpdatedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedBy

`func (o *AlertResponseProperties) SetUpdatedBy(v string)`

SetUpdatedBy sets UpdatedBy field to given value.

### HasUpdatedBy

`func (o *AlertResponseProperties) HasUpdatedBy() bool`

HasUpdatedBy returns a boolean if a field has been set.

### SetUpdatedByNil

`func (o *AlertResponseProperties) SetUpdatedByNil(b bool)`

 SetUpdatedByNil sets the value for UpdatedBy to be an explicit nil

### UnsetUpdatedBy
`func (o *AlertResponseProperties) UnsetUpdatedBy()`

UnsetUpdatedBy ensures that no value is present for UpdatedBy, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


