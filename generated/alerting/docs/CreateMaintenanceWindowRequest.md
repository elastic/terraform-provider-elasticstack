# CreateMaintenanceWindowRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Enabled** | Pointer to **bool** | Whether the current maintenance window is enabled. Disabled maintenance windows do not suppress notifications. | [optional] 
**Schedule** | [**CreateMaintenanceWindowRequestSchedule**](CreateMaintenanceWindowRequestSchedule.md) |  | 
**Scope** | Pointer to [**CreateMaintenanceWindowRequestScope**](CreateMaintenanceWindowRequestScope.md) |  | [optional] 
**Title** | **string** | The name of the maintenance window. While this name does not have to be unique, a distinctive name can help you identify a specific maintenance window. | 

## Methods

### NewCreateMaintenanceWindowRequest

`func NewCreateMaintenanceWindowRequest(schedule CreateMaintenanceWindowRequestSchedule, title string, ) *CreateMaintenanceWindowRequest`

NewCreateMaintenanceWindowRequest instantiates a new CreateMaintenanceWindowRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewCreateMaintenanceWindowRequestWithDefaults

`func NewCreateMaintenanceWindowRequestWithDefaults() *CreateMaintenanceWindowRequest`

NewCreateMaintenanceWindowRequestWithDefaults instantiates a new CreateMaintenanceWindowRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetEnabled

`func (o *CreateMaintenanceWindowRequest) GetEnabled() bool`

GetEnabled returns the Enabled field if non-nil, zero value otherwise.

### GetEnabledOk

`func (o *CreateMaintenanceWindowRequest) GetEnabledOk() (*bool, bool)`

GetEnabledOk returns a tuple with the Enabled field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnabled

`func (o *CreateMaintenanceWindowRequest) SetEnabled(v bool)`

SetEnabled sets Enabled field to given value.

### HasEnabled

`func (o *CreateMaintenanceWindowRequest) HasEnabled() bool`

HasEnabled returns a boolean if a field has been set.

### GetSchedule

`func (o *CreateMaintenanceWindowRequest) GetSchedule() CreateMaintenanceWindowRequestSchedule`

GetSchedule returns the Schedule field if non-nil, zero value otherwise.

### GetScheduleOk

`func (o *CreateMaintenanceWindowRequest) GetScheduleOk() (*CreateMaintenanceWindowRequestSchedule, bool)`

GetScheduleOk returns a tuple with the Schedule field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSchedule

`func (o *CreateMaintenanceWindowRequest) SetSchedule(v CreateMaintenanceWindowRequestSchedule)`

SetSchedule sets Schedule field to given value.


### GetScope

`func (o *CreateMaintenanceWindowRequest) GetScope() CreateMaintenanceWindowRequestScope`

GetScope returns the Scope field if non-nil, zero value otherwise.

### GetScopeOk

`func (o *CreateMaintenanceWindowRequest) GetScopeOk() (*CreateMaintenanceWindowRequestScope, bool)`

GetScopeOk returns a tuple with the Scope field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScope

`func (o *CreateMaintenanceWindowRequest) SetScope(v CreateMaintenanceWindowRequestScope)`

SetScope sets Scope field to given value.

### HasScope

`func (o *CreateMaintenanceWindowRequest) HasScope() bool`

HasScope returns a boolean if a field has been set.

### GetTitle

`func (o *CreateMaintenanceWindowRequest) GetTitle() string`

GetTitle returns the Title field if non-nil, zero value otherwise.

### GetTitleOk

`func (o *CreateMaintenanceWindowRequest) GetTitleOk() (*string, bool)`

GetTitleOk returns a tuple with the Title field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTitle

`func (o *CreateMaintenanceWindowRequest) SetTitle(v string)`

SetTitle sets Title field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


