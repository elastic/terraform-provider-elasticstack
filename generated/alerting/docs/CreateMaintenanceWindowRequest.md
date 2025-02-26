# CreateMaintenanceWindowRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Title** | **string** | The name of the maintenance window. While this name does not have to be unique, a distinctive name can help you identify a maintenance window.  | 
**Enabled** | Pointer to **bool** | Indicates whether the maintenance window is currently enabled. | [optional] 
**Duration** | **float32** |  | 
**Start** | **string** | An ISO date. | 

## Methods

### NewCreateMaintenanceWindowRequest

`func NewCreateMaintenanceWindowRequest(title string, duration float32, start string, ) *CreateMaintenanceWindowRequest`

NewCreateMaintenanceWindowRequest instantiates a new CreateMaintenanceWindowRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewCreateMaintenanceWindowRequestWithDefaults

`func NewCreateMaintenanceWindowRequestWithDefaults() *CreateMaintenanceWindowRequest`

NewCreateMaintenanceWindowRequestWithDefaults instantiates a new CreateMaintenanceWindowRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

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

### GetDuration

`func (o *CreateMaintenanceWindowRequest) GetDuration() float32`

GetDuration returns the Duration field if non-nil, zero value otherwise.

### GetDurationOk

`func (o *CreateMaintenanceWindowRequest) GetDurationOk() (*float32, bool)`

GetDurationOk returns a tuple with the Duration field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDuration

`func (o *CreateMaintenanceWindowRequest) SetDuration(v float32)`

SetDuration sets Duration field to given value.


### GetStart

`func (o *CreateMaintenanceWindowRequest) GetStart() string`

GetStart returns the Start field if non-nil, zero value otherwise.

### GetStartOk

`func (o *CreateMaintenanceWindowRequest) GetStartOk() (*string, bool)`

GetStartOk returns a tuple with the Start field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStart

`func (o *CreateMaintenanceWindowRequest) SetStart(v string)`

SetStart sets Start field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


