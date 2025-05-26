# PatchAlertingMaintenanceWindowIdRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Enabled** | Pointer to **interface{}** | Whether the current maintenance window is enabled. Disabled maintenance windows do not suppress notifications. | [optional] 
**Schedule** | Pointer to [**PostAlertingMaintenanceWindow200ResponseSchedule**](PostAlertingMaintenanceWindow200ResponseSchedule.md) |  | [optional] 
**Scope** | Pointer to [**PostAlertingMaintenanceWindowRequestScope**](PostAlertingMaintenanceWindowRequestScope.md) |  | [optional] 
**Title** | Pointer to **interface{}** | The name of the maintenance window. While this name does not have to be unique, a distinctive name can help you identify a specific maintenance window. | [optional] 

## Methods

### NewPatchAlertingMaintenanceWindowIdRequest

`func NewPatchAlertingMaintenanceWindowIdRequest() *PatchAlertingMaintenanceWindowIdRequest`

NewPatchAlertingMaintenanceWindowIdRequest instantiates a new PatchAlertingMaintenanceWindowIdRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPatchAlertingMaintenanceWindowIdRequestWithDefaults

`func NewPatchAlertingMaintenanceWindowIdRequestWithDefaults() *PatchAlertingMaintenanceWindowIdRequest`

NewPatchAlertingMaintenanceWindowIdRequestWithDefaults instantiates a new PatchAlertingMaintenanceWindowIdRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetEnabled

`func (o *PatchAlertingMaintenanceWindowIdRequest) GetEnabled() interface{}`

GetEnabled returns the Enabled field if non-nil, zero value otherwise.

### GetEnabledOk

`func (o *PatchAlertingMaintenanceWindowIdRequest) GetEnabledOk() (*interface{}, bool)`

GetEnabledOk returns a tuple with the Enabled field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnabled

`func (o *PatchAlertingMaintenanceWindowIdRequest) SetEnabled(v interface{})`

SetEnabled sets Enabled field to given value.

### HasEnabled

`func (o *PatchAlertingMaintenanceWindowIdRequest) HasEnabled() bool`

HasEnabled returns a boolean if a field has been set.

### SetEnabledNil

`func (o *PatchAlertingMaintenanceWindowIdRequest) SetEnabledNil(b bool)`

 SetEnabledNil sets the value for Enabled to be an explicit nil

### UnsetEnabled
`func (o *PatchAlertingMaintenanceWindowIdRequest) UnsetEnabled()`

UnsetEnabled ensures that no value is present for Enabled, not even an explicit nil
### GetSchedule

`func (o *PatchAlertingMaintenanceWindowIdRequest) GetSchedule() PostAlertingMaintenanceWindow200ResponseSchedule`

GetSchedule returns the Schedule field if non-nil, zero value otherwise.

### GetScheduleOk

`func (o *PatchAlertingMaintenanceWindowIdRequest) GetScheduleOk() (*PostAlertingMaintenanceWindow200ResponseSchedule, bool)`

GetScheduleOk returns a tuple with the Schedule field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSchedule

`func (o *PatchAlertingMaintenanceWindowIdRequest) SetSchedule(v PostAlertingMaintenanceWindow200ResponseSchedule)`

SetSchedule sets Schedule field to given value.

### HasSchedule

`func (o *PatchAlertingMaintenanceWindowIdRequest) HasSchedule() bool`

HasSchedule returns a boolean if a field has been set.

### GetScope

`func (o *PatchAlertingMaintenanceWindowIdRequest) GetScope() PostAlertingMaintenanceWindowRequestScope`

GetScope returns the Scope field if non-nil, zero value otherwise.

### GetScopeOk

`func (o *PatchAlertingMaintenanceWindowIdRequest) GetScopeOk() (*PostAlertingMaintenanceWindowRequestScope, bool)`

GetScopeOk returns a tuple with the Scope field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScope

`func (o *PatchAlertingMaintenanceWindowIdRequest) SetScope(v PostAlertingMaintenanceWindowRequestScope)`

SetScope sets Scope field to given value.

### HasScope

`func (o *PatchAlertingMaintenanceWindowIdRequest) HasScope() bool`

HasScope returns a boolean if a field has been set.

### GetTitle

`func (o *PatchAlertingMaintenanceWindowIdRequest) GetTitle() interface{}`

GetTitle returns the Title field if non-nil, zero value otherwise.

### GetTitleOk

`func (o *PatchAlertingMaintenanceWindowIdRequest) GetTitleOk() (*interface{}, bool)`

GetTitleOk returns a tuple with the Title field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTitle

`func (o *PatchAlertingMaintenanceWindowIdRequest) SetTitle(v interface{})`

SetTitle sets Title field to given value.

### HasTitle

`func (o *PatchAlertingMaintenanceWindowIdRequest) HasTitle() bool`

HasTitle returns a boolean if a field has been set.

### SetTitleNil

`func (o *PatchAlertingMaintenanceWindowIdRequest) SetTitleNil(b bool)`

 SetTitleNil sets the value for Title to be an explicit nil

### UnsetTitle
`func (o *PatchAlertingMaintenanceWindowIdRequest) UnsetTitle()`

UnsetTitle ensures that no value is present for Title, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


