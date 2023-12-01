# DeleteSloInstancesRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**List** | [**[]DeleteSloInstancesRequestListInner**](DeleteSloInstancesRequestListInner.md) | An array of slo id and instance id | 

## Methods

### NewDeleteSloInstancesRequest

`func NewDeleteSloInstancesRequest(list []DeleteSloInstancesRequestListInner, ) *DeleteSloInstancesRequest`

NewDeleteSloInstancesRequest instantiates a new DeleteSloInstancesRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDeleteSloInstancesRequestWithDefaults

`func NewDeleteSloInstancesRequestWithDefaults() *DeleteSloInstancesRequest`

NewDeleteSloInstancesRequestWithDefaults instantiates a new DeleteSloInstancesRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetList

`func (o *DeleteSloInstancesRequest) GetList() []DeleteSloInstancesRequestListInner`

GetList returns the List field if non-nil, zero value otherwise.

### GetListOk

`func (o *DeleteSloInstancesRequest) GetListOk() (*[]DeleteSloInstancesRequestListInner, bool)`

GetListOk returns a tuple with the List field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetList

`func (o *DeleteSloInstancesRequest) SetList(v []DeleteSloInstancesRequestListInner)`

SetList sets List field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


