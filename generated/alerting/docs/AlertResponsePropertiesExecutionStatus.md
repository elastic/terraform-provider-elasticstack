# AlertResponsePropertiesExecutionStatus

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**LastExecutionDate** | Pointer to **time.Time** |  | [optional] 
**Status** | Pointer to **string** |  | [optional] 

## Methods

### NewAlertResponsePropertiesExecutionStatus

`func NewAlertResponsePropertiesExecutionStatus() *AlertResponsePropertiesExecutionStatus`

NewAlertResponsePropertiesExecutionStatus instantiates a new AlertResponsePropertiesExecutionStatus object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAlertResponsePropertiesExecutionStatusWithDefaults

`func NewAlertResponsePropertiesExecutionStatusWithDefaults() *AlertResponsePropertiesExecutionStatus`

NewAlertResponsePropertiesExecutionStatusWithDefaults instantiates a new AlertResponsePropertiesExecutionStatus object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetLastExecutionDate

`func (o *AlertResponsePropertiesExecutionStatus) GetLastExecutionDate() time.Time`

GetLastExecutionDate returns the LastExecutionDate field if non-nil, zero value otherwise.

### GetLastExecutionDateOk

`func (o *AlertResponsePropertiesExecutionStatus) GetLastExecutionDateOk() (*time.Time, bool)`

GetLastExecutionDateOk returns a tuple with the LastExecutionDate field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLastExecutionDate

`func (o *AlertResponsePropertiesExecutionStatus) SetLastExecutionDate(v time.Time)`

SetLastExecutionDate sets LastExecutionDate field to given value.

### HasLastExecutionDate

`func (o *AlertResponsePropertiesExecutionStatus) HasLastExecutionDate() bool`

HasLastExecutionDate returns a boolean if a field has been set.

### GetStatus

`func (o *AlertResponsePropertiesExecutionStatus) GetStatus() string`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *AlertResponsePropertiesExecutionStatus) GetStatusOk() (*string, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *AlertResponsePropertiesExecutionStatus) SetStatus(v string)`

SetStatus sets Status field to given value.

### HasStatus

`func (o *AlertResponsePropertiesExecutionStatus) HasStatus() bool`

HasStatus returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


