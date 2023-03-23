# ConfigPropertiesIndex

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ExecutionTimeField** | Pointer to **NullableString** | Specifies a field that will contain the time the alert condition was detected. | [optional] 
**Index** | **string** | The Elasticsearch index to be written to. | 
**Refresh** | Pointer to **bool** | The refresh policy for the write request, which affects when changes are made visible to search. Refer to the refresh setting for Elasticsearch document APIs.  | [optional] [default to false]

## Methods

### NewConfigPropertiesIndex

`func NewConfigPropertiesIndex(index string, ) *ConfigPropertiesIndex`

NewConfigPropertiesIndex instantiates a new ConfigPropertiesIndex object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewConfigPropertiesIndexWithDefaults

`func NewConfigPropertiesIndexWithDefaults() *ConfigPropertiesIndex`

NewConfigPropertiesIndexWithDefaults instantiates a new ConfigPropertiesIndex object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetExecutionTimeField

`func (o *ConfigPropertiesIndex) GetExecutionTimeField() string`

GetExecutionTimeField returns the ExecutionTimeField field if non-nil, zero value otherwise.

### GetExecutionTimeFieldOk

`func (o *ConfigPropertiesIndex) GetExecutionTimeFieldOk() (*string, bool)`

GetExecutionTimeFieldOk returns a tuple with the ExecutionTimeField field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExecutionTimeField

`func (o *ConfigPropertiesIndex) SetExecutionTimeField(v string)`

SetExecutionTimeField sets ExecutionTimeField field to given value.

### HasExecutionTimeField

`func (o *ConfigPropertiesIndex) HasExecutionTimeField() bool`

HasExecutionTimeField returns a boolean if a field has been set.

### SetExecutionTimeFieldNil

`func (o *ConfigPropertiesIndex) SetExecutionTimeFieldNil(b bool)`

 SetExecutionTimeFieldNil sets the value for ExecutionTimeField to be an explicit nil

### UnsetExecutionTimeField
`func (o *ConfigPropertiesIndex) UnsetExecutionTimeField()`

UnsetExecutionTimeField ensures that no value is present for ExecutionTimeField, not even an explicit nil
### GetIndex

`func (o *ConfigPropertiesIndex) GetIndex() string`

GetIndex returns the Index field if non-nil, zero value otherwise.

### GetIndexOk

`func (o *ConfigPropertiesIndex) GetIndexOk() (*string, bool)`

GetIndexOk returns a tuple with the Index field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIndex

`func (o *ConfigPropertiesIndex) SetIndex(v string)`

SetIndex sets Index field to given value.


### GetRefresh

`func (o *ConfigPropertiesIndex) GetRefresh() bool`

GetRefresh returns the Refresh field if non-nil, zero value otherwise.

### GetRefreshOk

`func (o *ConfigPropertiesIndex) GetRefreshOk() (*bool, bool)`

GetRefreshOk returns a tuple with the Refresh field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRefresh

`func (o *ConfigPropertiesIndex) SetRefresh(v bool)`

SetRefresh sets Refresh field to given value.

### HasRefresh

`func (o *ConfigPropertiesIndex) HasRefresh() bool`

HasRefresh returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


