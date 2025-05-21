# PostAlertingMaintenanceWindowRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Enabled** | Pointer to **interface{}** | Whether the current maintenance window is enabled. Disabled maintenance windows do not suppress notifications. | [optional] 
**Schedule** | [**PostAlertingMaintenanceWindowRequestSchedule**](PostAlertingMaintenanceWindowRequestSchedule.md) |  | 
**Scope** | Pointer to [**PostAlertingMaintenanceWindowRequestScope**](PostAlertingMaintenanceWindowRequestScope.md) |  | [optional] 
**Title** | **interface{}** | The name of the maintenance window. While this name does not have to be unique, a distinctive name can help you identify a specific maintenance window. | 

## Methods

### NewPostAlertingMaintenanceWindowRequest

`func NewPostAlertingMaintenanceWindowRequest(schedule PostAlertingMaintenanceWindowRequestSchedule, title interface{}, ) *PostAlertingMaintenanceWindowRequest`

NewPostAlertingMaintenanceWindowRequest instantiates a new PostAlertingMaintenanceWindowRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPostAlertingMaintenanceWindowRequestWithDefaults

`func NewPostAlertingMaintenanceWindowRequestWithDefaults() *PostAlertingMaintenanceWindowRequest`

NewPostAlertingMaintenanceWindowRequestWithDefaults instantiates a new PostAlertingMaintenanceWindowRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetEnabled

`func (o *PostAlertingMaintenanceWindowRequest) GetEnabled() interface{}`

GetEnabled returns the Enabled field if non-nil, zero value otherwise.

### GetEnabledOk

`func (o *PostAlertingMaintenanceWindowRequest) GetEnabledOk() (*interface{}, bool)`

GetEnabledOk returns a tuple with the Enabled field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnabled

`func (o *PostAlertingMaintenanceWindowRequest) SetEnabled(v interface{})`

SetEnabled sets Enabled field to given value.

### HasEnabled

`func (o *PostAlertingMaintenanceWindowRequest) HasEnabled() bool`

HasEnabled returns a boolean if a field has been set.

### SetEnabledNil

`func (o *PostAlertingMaintenanceWindowRequest) SetEnabledNil(b bool)`

 SetEnabledNil sets the value for Enabled to be an explicit nil

### UnsetEnabled
`func (o *PostAlertingMaintenanceWindowRequest) UnsetEnabled()`

UnsetEnabled ensures that no value is present for Enabled, not even an explicit nil
### GetSchedule

`func (o *PostAlertingMaintenanceWindowRequest) GetSchedule() PostAlertingMaintenanceWindowRequestSchedule`

GetSchedule returns the Schedule field if non-nil, zero value otherwise.

### GetScheduleOk

`func (o *PostAlertingMaintenanceWindowRequest) GetScheduleOk() (*PostAlertingMaintenanceWindowRequestSchedule, bool)`

GetScheduleOk returns a tuple with the Schedule field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSchedule

`func (o *PostAlertingMaintenanceWindowRequest) SetSchedule(v PostAlertingMaintenanceWindowRequestSchedule)`

SetSchedule sets Schedule field to given value.


### GetScope

`func (o *PostAlertingMaintenanceWindowRequest) GetScope() PostAlertingMaintenanceWindowRequestScope`

GetScope returns the Scope field if non-nil, zero value otherwise.

### GetScopeOk

`func (o *PostAlertingMaintenanceWindowRequest) GetScopeOk() (*PostAlertingMaintenanceWindowRequestScope, bool)`

GetScopeOk returns a tuple with the Scope field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScope

`func (o *PostAlertingMaintenanceWindowRequest) SetScope(v PostAlertingMaintenanceWindowRequestScope)`

SetScope sets Scope field to given value.

### HasScope

`func (o *PostAlertingMaintenanceWindowRequest) HasScope() bool`

HasScope returns a boolean if a field has been set.

### GetTitle

`func (o *PostAlertingMaintenanceWindowRequest) GetTitle() interface{}`

GetTitle returns the Title field if non-nil, zero value otherwise.

### GetTitleOk

`func (o *PostAlertingMaintenanceWindowRequest) GetTitleOk() (*interface{}, bool)`

GetTitleOk returns a tuple with the Title field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTitle

`func (o *PostAlertingMaintenanceWindowRequest) SetTitle(v interface{})`

SetTitle sets Title field to given value.


### SetTitleNil

`func (o *PostAlertingMaintenanceWindowRequest) SetTitleNil(b bool)`

 SetTitleNil sets the value for Title to be an explicit nil

### UnsetTitle
`func (o *PostAlertingMaintenanceWindowRequest) UnsetTitle()`

UnsetTitle ensures that no value is present for Title, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


