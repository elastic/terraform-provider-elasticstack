# DeleteSloInstancesRequestListInner

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**SloId** | **string** | The SLO unique identifier | 
**InstanceId** | **string** | The SLO instance identifier | 

## Methods

### NewDeleteSloInstancesRequestListInner

`func NewDeleteSloInstancesRequestListInner(sloId string, instanceId string, ) *DeleteSloInstancesRequestListInner`

NewDeleteSloInstancesRequestListInner instantiates a new DeleteSloInstancesRequestListInner object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDeleteSloInstancesRequestListInnerWithDefaults

`func NewDeleteSloInstancesRequestListInnerWithDefaults() *DeleteSloInstancesRequestListInner`

NewDeleteSloInstancesRequestListInnerWithDefaults instantiates a new DeleteSloInstancesRequestListInner object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetSloId

`func (o *DeleteSloInstancesRequestListInner) GetSloId() string`

GetSloId returns the SloId field if non-nil, zero value otherwise.

### GetSloIdOk

`func (o *DeleteSloInstancesRequestListInner) GetSloIdOk() (*string, bool)`

GetSloIdOk returns a tuple with the SloId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSloId

`func (o *DeleteSloInstancesRequestListInner) SetSloId(v string)`

SetSloId sets SloId field to given value.


### GetInstanceId

`func (o *DeleteSloInstancesRequestListInner) GetInstanceId() string`

GetInstanceId returns the InstanceId field if non-nil, zero value otherwise.

### GetInstanceIdOk

`func (o *DeleteSloInstancesRequestListInner) GetInstanceIdOk() (*string, bool)`

GetInstanceIdOk returns a tuple with the InstanceId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInstanceId

`func (o *DeleteSloInstancesRequestListInner) SetInstanceId(v string)`

SetInstanceId sets InstanceId field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


