# MaintenanceWindowResponseProperties

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**CreatedAt** | **string** | The date and time when the maintenance window was created. | 
**CreatedBy** | **string** | The identifier for the user that created the maintenance window. | 
**Enabled** | **bool** | Whether the current maintenance window is enabled. Disabled maintenance windows do not suppress notifications. | 
**Id** | **string** | The identifier for the maintenance window. | 
**Schedule** | [**MaintenanceWindowResponsePropertiesSchedule**](MaintenanceWindowResponsePropertiesSchedule.md) |  | 
**Scope** | Pointer to [**MaintenanceWindowResponsePropertiesScope**](MaintenanceWindowResponsePropertiesScope.md) |  | [optional] 
**Status** | **string** | The current status of the maintenance window. | 
**Title** | **string** | The name of the maintenance window. | 
**UpdatedAt** | **string** | The date and time when the maintenance window was last updated. | 
**UpdatedBy** | **string** | The identifier for the user that last updated this maintenance window. | 

## Methods

### NewMaintenanceWindowResponseProperties

`func NewMaintenanceWindowResponseProperties(createdAt string, createdBy string, enabled bool, id string, schedule MaintenanceWindowResponsePropertiesSchedule, status string, title string, updatedAt string, updatedBy string, ) *MaintenanceWindowResponseProperties`

NewMaintenanceWindowResponseProperties instantiates a new MaintenanceWindowResponseProperties object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewMaintenanceWindowResponsePropertiesWithDefaults

`func NewMaintenanceWindowResponsePropertiesWithDefaults() *MaintenanceWindowResponseProperties`

NewMaintenanceWindowResponsePropertiesWithDefaults instantiates a new MaintenanceWindowResponseProperties object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCreatedAt

`func (o *MaintenanceWindowResponseProperties) GetCreatedAt() string`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *MaintenanceWindowResponseProperties) GetCreatedAtOk() (*string, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *MaintenanceWindowResponseProperties) SetCreatedAt(v string)`

SetCreatedAt sets CreatedAt field to given value.


### GetCreatedBy

`func (o *MaintenanceWindowResponseProperties) GetCreatedBy() string`

GetCreatedBy returns the CreatedBy field if non-nil, zero value otherwise.

### GetCreatedByOk

`func (o *MaintenanceWindowResponseProperties) GetCreatedByOk() (*string, bool)`

GetCreatedByOk returns a tuple with the CreatedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedBy

`func (o *MaintenanceWindowResponseProperties) SetCreatedBy(v string)`

SetCreatedBy sets CreatedBy field to given value.


### GetEnabled

`func (o *MaintenanceWindowResponseProperties) GetEnabled() bool`

GetEnabled returns the Enabled field if non-nil, zero value otherwise.

### GetEnabledOk

`func (o *MaintenanceWindowResponseProperties) GetEnabledOk() (*bool, bool)`

GetEnabledOk returns a tuple with the Enabled field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnabled

`func (o *MaintenanceWindowResponseProperties) SetEnabled(v bool)`

SetEnabled sets Enabled field to given value.


### GetId

`func (o *MaintenanceWindowResponseProperties) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *MaintenanceWindowResponseProperties) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *MaintenanceWindowResponseProperties) SetId(v string)`

SetId sets Id field to given value.


### GetSchedule

`func (o *MaintenanceWindowResponseProperties) GetSchedule() MaintenanceWindowResponsePropertiesSchedule`

GetSchedule returns the Schedule field if non-nil, zero value otherwise.

### GetScheduleOk

`func (o *MaintenanceWindowResponseProperties) GetScheduleOk() (*MaintenanceWindowResponsePropertiesSchedule, bool)`

GetScheduleOk returns a tuple with the Schedule field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSchedule

`func (o *MaintenanceWindowResponseProperties) SetSchedule(v MaintenanceWindowResponsePropertiesSchedule)`

SetSchedule sets Schedule field to given value.


### GetScope

`func (o *MaintenanceWindowResponseProperties) GetScope() MaintenanceWindowResponsePropertiesScope`

GetScope returns the Scope field if non-nil, zero value otherwise.

### GetScopeOk

`func (o *MaintenanceWindowResponseProperties) GetScopeOk() (*MaintenanceWindowResponsePropertiesScope, bool)`

GetScopeOk returns a tuple with the Scope field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScope

`func (o *MaintenanceWindowResponseProperties) SetScope(v MaintenanceWindowResponsePropertiesScope)`

SetScope sets Scope field to given value.

### HasScope

`func (o *MaintenanceWindowResponseProperties) HasScope() bool`

HasScope returns a boolean if a field has been set.

### GetStatus

`func (o *MaintenanceWindowResponseProperties) GetStatus() string`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *MaintenanceWindowResponseProperties) GetStatusOk() (*string, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *MaintenanceWindowResponseProperties) SetStatus(v string)`

SetStatus sets Status field to given value.


### GetTitle

`func (o *MaintenanceWindowResponseProperties) GetTitle() string`

GetTitle returns the Title field if non-nil, zero value otherwise.

### GetTitleOk

`func (o *MaintenanceWindowResponseProperties) GetTitleOk() (*string, bool)`

GetTitleOk returns a tuple with the Title field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTitle

`func (o *MaintenanceWindowResponseProperties) SetTitle(v string)`

SetTitle sets Title field to given value.


### GetUpdatedAt

`func (o *MaintenanceWindowResponseProperties) GetUpdatedAt() string`

GetUpdatedAt returns the UpdatedAt field if non-nil, zero value otherwise.

### GetUpdatedAtOk

`func (o *MaintenanceWindowResponseProperties) GetUpdatedAtOk() (*string, bool)`

GetUpdatedAtOk returns a tuple with the UpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedAt

`func (o *MaintenanceWindowResponseProperties) SetUpdatedAt(v string)`

SetUpdatedAt sets UpdatedAt field to given value.


### GetUpdatedBy

`func (o *MaintenanceWindowResponseProperties) GetUpdatedBy() string`

GetUpdatedBy returns the UpdatedBy field if non-nil, zero value otherwise.

### GetUpdatedByOk

`func (o *MaintenanceWindowResponseProperties) GetUpdatedByOk() (*string, bool)`

GetUpdatedByOk returns a tuple with the UpdatedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedBy

`func (o *MaintenanceWindowResponseProperties) SetUpdatedBy(v string)`

SetUpdatedBy sets UpdatedBy field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


