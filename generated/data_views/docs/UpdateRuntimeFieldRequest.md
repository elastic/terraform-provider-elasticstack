# UpdateRuntimeFieldRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**RuntimeField** | **map[string]interface{}** | The runtime field definition object.  You can update following fields:  - &#x60;type&#x60; - &#x60;script&#x60;  | 

## Methods

### NewUpdateRuntimeFieldRequest

`func NewUpdateRuntimeFieldRequest(runtimeField map[string]interface{}, ) *UpdateRuntimeFieldRequest`

NewUpdateRuntimeFieldRequest instantiates a new UpdateRuntimeFieldRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewUpdateRuntimeFieldRequestWithDefaults

`func NewUpdateRuntimeFieldRequestWithDefaults() *UpdateRuntimeFieldRequest`

NewUpdateRuntimeFieldRequestWithDefaults instantiates a new UpdateRuntimeFieldRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetRuntimeField

`func (o *UpdateRuntimeFieldRequest) GetRuntimeField() map[string]interface{}`

GetRuntimeField returns the RuntimeField field if non-nil, zero value otherwise.

### GetRuntimeFieldOk

`func (o *UpdateRuntimeFieldRequest) GetRuntimeFieldOk() (*map[string]interface{}, bool)`

GetRuntimeFieldOk returns a tuple with the RuntimeField field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRuntimeField

`func (o *UpdateRuntimeFieldRequest) SetRuntimeField(v map[string]interface{})`

SetRuntimeField sets RuntimeField field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


