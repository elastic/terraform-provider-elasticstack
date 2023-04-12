# RunConnector200Response

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ConnectorId** | **string** | The identifier for the connector. | 
**Data** | Pointer to [**RunConnector200ResponseData**](RunConnector200ResponseData.md) |  | [optional] 
**Status** | **string** | The status of the action. | 

## Methods

### NewRunConnector200Response

`func NewRunConnector200Response(connectorId string, status string, ) *RunConnector200Response`

NewRunConnector200Response instantiates a new RunConnector200Response object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewRunConnector200ResponseWithDefaults

`func NewRunConnector200ResponseWithDefaults() *RunConnector200Response`

NewRunConnector200ResponseWithDefaults instantiates a new RunConnector200Response object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetConnectorId

`func (o *RunConnector200Response) GetConnectorId() string`

GetConnectorId returns the ConnectorId field if non-nil, zero value otherwise.

### GetConnectorIdOk

`func (o *RunConnector200Response) GetConnectorIdOk() (*string, bool)`

GetConnectorIdOk returns a tuple with the ConnectorId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConnectorId

`func (o *RunConnector200Response) SetConnectorId(v string)`

SetConnectorId sets ConnectorId field to given value.


### GetData

`func (o *RunConnector200Response) GetData() RunConnector200ResponseData`

GetData returns the Data field if non-nil, zero value otherwise.

### GetDataOk

`func (o *RunConnector200Response) GetDataOk() (*RunConnector200ResponseData, bool)`

GetDataOk returns a tuple with the Data field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetData

`func (o *RunConnector200Response) SetData(v RunConnector200ResponseData)`

SetData sets Data field to given value.

### HasData

`func (o *RunConnector200Response) HasData() bool`

HasData returns a boolean if a field has been set.

### GetStatus

`func (o *RunConnector200Response) GetStatus() string`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *RunConnector200Response) GetStatusOk() (*string, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *RunConnector200Response) SetStatus(v string)`

SetStatus sets Status field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


