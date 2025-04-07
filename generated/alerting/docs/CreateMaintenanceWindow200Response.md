# CreateMaintenanceWindow200Response

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**CreatedAt** | **interface{}** | The date and time when the maintenance window was created. | 
**CreatedBy** | **interface{}** | The identifier for the user that created the maintenance window. | 
**Enabled** | **interface{}** | Whether the current maintenance window is enabled. Disabled maintenance windows do not suppress notifications. | 
**Id** | **interface{}** | The identifier for the maintenance window. | 
**Schedule** | [**CreateMaintenanceWindow200ResponseSchedule**](CreateMaintenanceWindow200ResponseSchedule.md) |  | 
**Scope** | Pointer to [**CreateMaintenanceWindowRequestScope**](CreateMaintenanceWindowRequestScope.md) |  | [optional] 
**Status** | **interface{}** | The current status of the maintenance window. | 
**Title** | **interface{}** | The name of the maintenance window. | 
**UpdatedAt** | **interface{}** | The date and time when the maintenance window was last updated. | 
**UpdatedBy** | **interface{}** | The identifier for the user that last updated this maintenance window. | 

## Methods

### NewCreateMaintenanceWindow200Response

`func NewCreateMaintenanceWindow200Response(createdAt interface{}, createdBy interface{}, enabled interface{}, id interface{}, schedule CreateMaintenanceWindow200ResponseSchedule, status interface{}, title interface{}, updatedAt interface{}, updatedBy interface{}, ) *CreateMaintenanceWindow200Response`

NewCreateMaintenanceWindow200Response instantiates a new CreateMaintenanceWindow200Response object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewCreateMaintenanceWindow200ResponseWithDefaults

`func NewCreateMaintenanceWindow200ResponseWithDefaults() *CreateMaintenanceWindow200Response`

NewCreateMaintenanceWindow200ResponseWithDefaults instantiates a new CreateMaintenanceWindow200Response object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCreatedAt

`func (o *CreateMaintenanceWindow200Response) GetCreatedAt() interface{}`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *CreateMaintenanceWindow200Response) GetCreatedAtOk() (*interface{}, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *CreateMaintenanceWindow200Response) SetCreatedAt(v interface{})`

SetCreatedAt sets CreatedAt field to given value.


### SetCreatedAtNil

`func (o *CreateMaintenanceWindow200Response) SetCreatedAtNil(b bool)`

 SetCreatedAtNil sets the value for CreatedAt to be an explicit nil

### UnsetCreatedAt
`func (o *CreateMaintenanceWindow200Response) UnsetCreatedAt()`

UnsetCreatedAt ensures that no value is present for CreatedAt, not even an explicit nil
### GetCreatedBy

`func (o *CreateMaintenanceWindow200Response) GetCreatedBy() interface{}`

GetCreatedBy returns the CreatedBy field if non-nil, zero value otherwise.

### GetCreatedByOk

`func (o *CreateMaintenanceWindow200Response) GetCreatedByOk() (*interface{}, bool)`

GetCreatedByOk returns a tuple with the CreatedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedBy

`func (o *CreateMaintenanceWindow200Response) SetCreatedBy(v interface{})`

SetCreatedBy sets CreatedBy field to given value.


### SetCreatedByNil

`func (o *CreateMaintenanceWindow200Response) SetCreatedByNil(b bool)`

 SetCreatedByNil sets the value for CreatedBy to be an explicit nil

### UnsetCreatedBy
`func (o *CreateMaintenanceWindow200Response) UnsetCreatedBy()`

UnsetCreatedBy ensures that no value is present for CreatedBy, not even an explicit nil
### GetEnabled

`func (o *CreateMaintenanceWindow200Response) GetEnabled() interface{}`

GetEnabled returns the Enabled field if non-nil, zero value otherwise.

### GetEnabledOk

`func (o *CreateMaintenanceWindow200Response) GetEnabledOk() (*interface{}, bool)`

GetEnabledOk returns a tuple with the Enabled field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnabled

`func (o *CreateMaintenanceWindow200Response) SetEnabled(v interface{})`

SetEnabled sets Enabled field to given value.


### SetEnabledNil

`func (o *CreateMaintenanceWindow200Response) SetEnabledNil(b bool)`

 SetEnabledNil sets the value for Enabled to be an explicit nil

### UnsetEnabled
`func (o *CreateMaintenanceWindow200Response) UnsetEnabled()`

UnsetEnabled ensures that no value is present for Enabled, not even an explicit nil
### GetId

`func (o *CreateMaintenanceWindow200Response) GetId() interface{}`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *CreateMaintenanceWindow200Response) GetIdOk() (*interface{}, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *CreateMaintenanceWindow200Response) SetId(v interface{})`

SetId sets Id field to given value.


### SetIdNil

`func (o *CreateMaintenanceWindow200Response) SetIdNil(b bool)`

 SetIdNil sets the value for Id to be an explicit nil

### UnsetId
`func (o *CreateMaintenanceWindow200Response) UnsetId()`

UnsetId ensures that no value is present for Id, not even an explicit nil
### GetSchedule

`func (o *CreateMaintenanceWindow200Response) GetSchedule() CreateMaintenanceWindow200ResponseSchedule`

GetSchedule returns the Schedule field if non-nil, zero value otherwise.

### GetScheduleOk

`func (o *CreateMaintenanceWindow200Response) GetScheduleOk() (*CreateMaintenanceWindow200ResponseSchedule, bool)`

GetScheduleOk returns a tuple with the Schedule field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSchedule

`func (o *CreateMaintenanceWindow200Response) SetSchedule(v CreateMaintenanceWindow200ResponseSchedule)`

SetSchedule sets Schedule field to given value.


### GetScope

`func (o *CreateMaintenanceWindow200Response) GetScope() CreateMaintenanceWindowRequestScope`

GetScope returns the Scope field if non-nil, zero value otherwise.

### GetScopeOk

`func (o *CreateMaintenanceWindow200Response) GetScopeOk() (*CreateMaintenanceWindowRequestScope, bool)`

GetScopeOk returns a tuple with the Scope field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScope

`func (o *CreateMaintenanceWindow200Response) SetScope(v CreateMaintenanceWindowRequestScope)`

SetScope sets Scope field to given value.

### HasScope

`func (o *CreateMaintenanceWindow200Response) HasScope() bool`

HasScope returns a boolean if a field has been set.

### GetStatus

`func (o *CreateMaintenanceWindow200Response) GetStatus() interface{}`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *CreateMaintenanceWindow200Response) GetStatusOk() (*interface{}, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *CreateMaintenanceWindow200Response) SetStatus(v interface{})`

SetStatus sets Status field to given value.


### SetStatusNil

`func (o *CreateMaintenanceWindow200Response) SetStatusNil(b bool)`

 SetStatusNil sets the value for Status to be an explicit nil

### UnsetStatus
`func (o *CreateMaintenanceWindow200Response) UnsetStatus()`

UnsetStatus ensures that no value is present for Status, not even an explicit nil
### GetTitle

`func (o *CreateMaintenanceWindow200Response) GetTitle() interface{}`

GetTitle returns the Title field if non-nil, zero value otherwise.

### GetTitleOk

`func (o *CreateMaintenanceWindow200Response) GetTitleOk() (*interface{}, bool)`

GetTitleOk returns a tuple with the Title field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTitle

`func (o *CreateMaintenanceWindow200Response) SetTitle(v interface{})`

SetTitle sets Title field to given value.


### SetTitleNil

`func (o *CreateMaintenanceWindow200Response) SetTitleNil(b bool)`

 SetTitleNil sets the value for Title to be an explicit nil

### UnsetTitle
`func (o *CreateMaintenanceWindow200Response) UnsetTitle()`

UnsetTitle ensures that no value is present for Title, not even an explicit nil
### GetUpdatedAt

`func (o *CreateMaintenanceWindow200Response) GetUpdatedAt() interface{}`

GetUpdatedAt returns the UpdatedAt field if non-nil, zero value otherwise.

### GetUpdatedAtOk

`func (o *CreateMaintenanceWindow200Response) GetUpdatedAtOk() (*interface{}, bool)`

GetUpdatedAtOk returns a tuple with the UpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedAt

`func (o *CreateMaintenanceWindow200Response) SetUpdatedAt(v interface{})`

SetUpdatedAt sets UpdatedAt field to given value.


### SetUpdatedAtNil

`func (o *CreateMaintenanceWindow200Response) SetUpdatedAtNil(b bool)`

 SetUpdatedAtNil sets the value for UpdatedAt to be an explicit nil

### UnsetUpdatedAt
`func (o *CreateMaintenanceWindow200Response) UnsetUpdatedAt()`

UnsetUpdatedAt ensures that no value is present for UpdatedAt, not even an explicit nil
### GetUpdatedBy

`func (o *CreateMaintenanceWindow200Response) GetUpdatedBy() interface{}`

GetUpdatedBy returns the UpdatedBy field if non-nil, zero value otherwise.

### GetUpdatedByOk

`func (o *CreateMaintenanceWindow200Response) GetUpdatedByOk() (*interface{}, bool)`

GetUpdatedByOk returns a tuple with the UpdatedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedBy

`func (o *CreateMaintenanceWindow200Response) SetUpdatedBy(v interface{})`

SetUpdatedBy sets UpdatedBy field to given value.


### SetUpdatedByNil

`func (o *CreateMaintenanceWindow200Response) SetUpdatedByNil(b bool)`

 SetUpdatedByNil sets the value for UpdatedBy to be an explicit nil

### UnsetUpdatedBy
`func (o *CreateMaintenanceWindow200Response) UnsetUpdatedBy()`

UnsetUpdatedBy ensures that no value is present for UpdatedBy, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


