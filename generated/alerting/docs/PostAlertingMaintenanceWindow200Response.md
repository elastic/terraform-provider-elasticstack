# PostAlertingMaintenanceWindow200Response

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**CreatedAt** | **interface{}** | The date and time when the maintenance window was created. | 
**CreatedBy** | **interface{}** | The identifier for the user that created the maintenance window. | 
**Enabled** | **interface{}** | Whether the current maintenance window is enabled. Disabled maintenance windows do not suppress notifications. | 
**Id** | **interface{}** | The identifier for the maintenance window. | 
**Schedule** | [**PostAlertingMaintenanceWindow200ResponseSchedule**](PostAlertingMaintenanceWindow200ResponseSchedule.md) |  | 
**Scope** | Pointer to [**PostAlertingMaintenanceWindowRequestScope**](PostAlertingMaintenanceWindowRequestScope.md) |  | [optional] 
**Status** | **interface{}** | The current status of the maintenance window. | 
**Title** | **interface{}** | The name of the maintenance window. | 
**UpdatedAt** | **interface{}** | The date and time when the maintenance window was last updated. | 
**UpdatedBy** | **interface{}** | The identifier for the user that last updated this maintenance window. | 

## Methods

### NewPostAlertingMaintenanceWindow200Response

`func NewPostAlertingMaintenanceWindow200Response(createdAt interface{}, createdBy interface{}, enabled interface{}, id interface{}, schedule PostAlertingMaintenanceWindow200ResponseSchedule, status interface{}, title interface{}, updatedAt interface{}, updatedBy interface{}, ) *PostAlertingMaintenanceWindow200Response`

NewPostAlertingMaintenanceWindow200Response instantiates a new PostAlertingMaintenanceWindow200Response object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPostAlertingMaintenanceWindow200ResponseWithDefaults

`func NewPostAlertingMaintenanceWindow200ResponseWithDefaults() *PostAlertingMaintenanceWindow200Response`

NewPostAlertingMaintenanceWindow200ResponseWithDefaults instantiates a new PostAlertingMaintenanceWindow200Response object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCreatedAt

`func (o *PostAlertingMaintenanceWindow200Response) GetCreatedAt() interface{}`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *PostAlertingMaintenanceWindow200Response) GetCreatedAtOk() (*interface{}, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *PostAlertingMaintenanceWindow200Response) SetCreatedAt(v interface{})`

SetCreatedAt sets CreatedAt field to given value.


### SetCreatedAtNil

`func (o *PostAlertingMaintenanceWindow200Response) SetCreatedAtNil(b bool)`

 SetCreatedAtNil sets the value for CreatedAt to be an explicit nil

### UnsetCreatedAt
`func (o *PostAlertingMaintenanceWindow200Response) UnsetCreatedAt()`

UnsetCreatedAt ensures that no value is present for CreatedAt, not even an explicit nil
### GetCreatedBy

`func (o *PostAlertingMaintenanceWindow200Response) GetCreatedBy() interface{}`

GetCreatedBy returns the CreatedBy field if non-nil, zero value otherwise.

### GetCreatedByOk

`func (o *PostAlertingMaintenanceWindow200Response) GetCreatedByOk() (*interface{}, bool)`

GetCreatedByOk returns a tuple with the CreatedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedBy

`func (o *PostAlertingMaintenanceWindow200Response) SetCreatedBy(v interface{})`

SetCreatedBy sets CreatedBy field to given value.


### SetCreatedByNil

`func (o *PostAlertingMaintenanceWindow200Response) SetCreatedByNil(b bool)`

 SetCreatedByNil sets the value for CreatedBy to be an explicit nil

### UnsetCreatedBy
`func (o *PostAlertingMaintenanceWindow200Response) UnsetCreatedBy()`

UnsetCreatedBy ensures that no value is present for CreatedBy, not even an explicit nil
### GetEnabled

`func (o *PostAlertingMaintenanceWindow200Response) GetEnabled() interface{}`

GetEnabled returns the Enabled field if non-nil, zero value otherwise.

### GetEnabledOk

`func (o *PostAlertingMaintenanceWindow200Response) GetEnabledOk() (*interface{}, bool)`

GetEnabledOk returns a tuple with the Enabled field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnabled

`func (o *PostAlertingMaintenanceWindow200Response) SetEnabled(v interface{})`

SetEnabled sets Enabled field to given value.


### SetEnabledNil

`func (o *PostAlertingMaintenanceWindow200Response) SetEnabledNil(b bool)`

 SetEnabledNil sets the value for Enabled to be an explicit nil

### UnsetEnabled
`func (o *PostAlertingMaintenanceWindow200Response) UnsetEnabled()`

UnsetEnabled ensures that no value is present for Enabled, not even an explicit nil
### GetId

`func (o *PostAlertingMaintenanceWindow200Response) GetId() interface{}`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *PostAlertingMaintenanceWindow200Response) GetIdOk() (*interface{}, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *PostAlertingMaintenanceWindow200Response) SetId(v interface{})`

SetId sets Id field to given value.


### SetIdNil

`func (o *PostAlertingMaintenanceWindow200Response) SetIdNil(b bool)`

 SetIdNil sets the value for Id to be an explicit nil

### UnsetId
`func (o *PostAlertingMaintenanceWindow200Response) UnsetId()`

UnsetId ensures that no value is present for Id, not even an explicit nil
### GetSchedule

`func (o *PostAlertingMaintenanceWindow200Response) GetSchedule() PostAlertingMaintenanceWindow200ResponseSchedule`

GetSchedule returns the Schedule field if non-nil, zero value otherwise.

### GetScheduleOk

`func (o *PostAlertingMaintenanceWindow200Response) GetScheduleOk() (*PostAlertingMaintenanceWindow200ResponseSchedule, bool)`

GetScheduleOk returns a tuple with the Schedule field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSchedule

`func (o *PostAlertingMaintenanceWindow200Response) SetSchedule(v PostAlertingMaintenanceWindow200ResponseSchedule)`

SetSchedule sets Schedule field to given value.


### GetScope

`func (o *PostAlertingMaintenanceWindow200Response) GetScope() PostAlertingMaintenanceWindowRequestScope`

GetScope returns the Scope field if non-nil, zero value otherwise.

### GetScopeOk

`func (o *PostAlertingMaintenanceWindow200Response) GetScopeOk() (*PostAlertingMaintenanceWindowRequestScope, bool)`

GetScopeOk returns a tuple with the Scope field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScope

`func (o *PostAlertingMaintenanceWindow200Response) SetScope(v PostAlertingMaintenanceWindowRequestScope)`

SetScope sets Scope field to given value.

### HasScope

`func (o *PostAlertingMaintenanceWindow200Response) HasScope() bool`

HasScope returns a boolean if a field has been set.

### GetStatus

`func (o *PostAlertingMaintenanceWindow200Response) GetStatus() interface{}`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *PostAlertingMaintenanceWindow200Response) GetStatusOk() (*interface{}, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *PostAlertingMaintenanceWindow200Response) SetStatus(v interface{})`

SetStatus sets Status field to given value.


### SetStatusNil

`func (o *PostAlertingMaintenanceWindow200Response) SetStatusNil(b bool)`

 SetStatusNil sets the value for Status to be an explicit nil

### UnsetStatus
`func (o *PostAlertingMaintenanceWindow200Response) UnsetStatus()`

UnsetStatus ensures that no value is present for Status, not even an explicit nil
### GetTitle

`func (o *PostAlertingMaintenanceWindow200Response) GetTitle() interface{}`

GetTitle returns the Title field if non-nil, zero value otherwise.

### GetTitleOk

`func (o *PostAlertingMaintenanceWindow200Response) GetTitleOk() (*interface{}, bool)`

GetTitleOk returns a tuple with the Title field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTitle

`func (o *PostAlertingMaintenanceWindow200Response) SetTitle(v interface{})`

SetTitle sets Title field to given value.


### SetTitleNil

`func (o *PostAlertingMaintenanceWindow200Response) SetTitleNil(b bool)`

 SetTitleNil sets the value for Title to be an explicit nil

### UnsetTitle
`func (o *PostAlertingMaintenanceWindow200Response) UnsetTitle()`

UnsetTitle ensures that no value is present for Title, not even an explicit nil
### GetUpdatedAt

`func (o *PostAlertingMaintenanceWindow200Response) GetUpdatedAt() interface{}`

GetUpdatedAt returns the UpdatedAt field if non-nil, zero value otherwise.

### GetUpdatedAtOk

`func (o *PostAlertingMaintenanceWindow200Response) GetUpdatedAtOk() (*interface{}, bool)`

GetUpdatedAtOk returns a tuple with the UpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedAt

`func (o *PostAlertingMaintenanceWindow200Response) SetUpdatedAt(v interface{})`

SetUpdatedAt sets UpdatedAt field to given value.


### SetUpdatedAtNil

`func (o *PostAlertingMaintenanceWindow200Response) SetUpdatedAtNil(b bool)`

 SetUpdatedAtNil sets the value for UpdatedAt to be an explicit nil

### UnsetUpdatedAt
`func (o *PostAlertingMaintenanceWindow200Response) UnsetUpdatedAt()`

UnsetUpdatedAt ensures that no value is present for UpdatedAt, not even an explicit nil
### GetUpdatedBy

`func (o *PostAlertingMaintenanceWindow200Response) GetUpdatedBy() interface{}`

GetUpdatedBy returns the UpdatedBy field if non-nil, zero value otherwise.

### GetUpdatedByOk

`func (o *PostAlertingMaintenanceWindow200Response) GetUpdatedByOk() (*interface{}, bool)`

GetUpdatedByOk returns a tuple with the UpdatedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedBy

`func (o *PostAlertingMaintenanceWindow200Response) SetUpdatedBy(v interface{})`

SetUpdatedBy sets UpdatedBy field to given value.


### SetUpdatedByNil

`func (o *PostAlertingMaintenanceWindow200Response) SetUpdatedByNil(b bool)`

 SetUpdatedByNil sets the value for UpdatedBy to be an explicit nil

### UnsetUpdatedBy
`func (o *PostAlertingMaintenanceWindow200Response) UnsetUpdatedBy()`

UnsetUpdatedBy ensures that no value is present for UpdatedBy, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


