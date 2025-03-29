# MaintenanceWindowResponseProperties

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **string** | The identifier for the maintenance window. | 
**Title** | **string** | The name of the maintenance window. | 
**Start** | **string** | The start date of the maintenance window. | 
**Duration** | **float32** | The duration of the maintenance window. | 
**Enabled** | **bool** | Indicates whether the maintenance window is currently enabled. | 
**CreatedBy** | **string** | The identifier for the user that created the maintenance window. | 
**CreatedAt** | **time.Time** | The date and time in which the maintenance window was created. | 
**UpdatedBy** | **string** | The identifier for the user that updated this maintenance window most recently. | 
**UpdatedAt** | **string** | The date and time that the maintenance window was updated most recently. | 
**Status** | **string** | The status of the maintenance window. One of the following values &#x60;running&#x60;, &#x60;upcoming&#x60;, &#x60;finished&#x60; or &#x60;archived&#x60;. | 

## Methods

### NewMaintenanceWindowResponseProperties

`func NewMaintenanceWindowResponseProperties(id string, title string, start string, duration float32, enabled bool, createdBy string, createdAt time.Time, updatedBy string, updatedAt string, status string, ) *MaintenanceWindowResponseProperties`

NewMaintenanceWindowResponseProperties instantiates a new MaintenanceWindowResponseProperties object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewMaintenanceWindowResponsePropertiesWithDefaults

`func NewMaintenanceWindowResponsePropertiesWithDefaults() *MaintenanceWindowResponseProperties`

NewMaintenanceWindowResponsePropertiesWithDefaults instantiates a new MaintenanceWindowResponseProperties object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

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


### GetStart

`func (o *MaintenanceWindowResponseProperties) GetStart() string`

GetStart returns the Start field if non-nil, zero value otherwise.

### GetStartOk

`func (o *MaintenanceWindowResponseProperties) GetStartOk() (*string, bool)`

GetStartOk returns a tuple with the Start field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStart

`func (o *MaintenanceWindowResponseProperties) SetStart(v string)`

SetStart sets Start field to given value.


### GetDuration

`func (o *MaintenanceWindowResponseProperties) GetDuration() float32`

GetDuration returns the Duration field if non-nil, zero value otherwise.

### GetDurationOk

`func (o *MaintenanceWindowResponseProperties) GetDurationOk() (*float32, bool)`

GetDurationOk returns a tuple with the Duration field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDuration

`func (o *MaintenanceWindowResponseProperties) SetDuration(v float32)`

SetDuration sets Duration field to given value.


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


### GetCreatedAt

`func (o *MaintenanceWindowResponseProperties) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *MaintenanceWindowResponseProperties) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *MaintenanceWindowResponseProperties) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.


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



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


