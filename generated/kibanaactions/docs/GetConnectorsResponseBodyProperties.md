# GetConnectorsResponseBodyProperties

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ConnectorTypeId** | [**ConnectorTypes**](ConnectorTypes.md) |  | 
**Config** | Pointer to **map[string]interface{}** | The configuration for the connector. Configuration properties vary depending on the connector type. | [optional] 
**Id** | **string** | The identifier for the connector. | 
**IsDeprecated** | **bool** | Indicates whether the connector type is deprecated. | 
**IsMissingSecrets** | Pointer to **bool** | Indicates whether secrets are missing for the connector. Secrets configuration properties vary depending on the connector type. | [optional] 
**IsPreconfigured** | **bool** | Indicates whether it is a preconfigured connector. If true, the &#x60;config&#x60; and &#x60;is_missing_secrets&#x60; properties are omitted from the response. | 
**Name** | **string** | The display name for the connector. | 
**ReferencedByCount** | **int32** | Indicates the number of saved objects that reference the connector. If &#x60;is_preconfigured&#x60; is true, this value is not calculated. | [default to 0]

## Methods

### NewGetConnectorsResponseBodyProperties

`func NewGetConnectorsResponseBodyProperties(connectorTypeId ConnectorTypes, id string, isDeprecated bool, isPreconfigured bool, name string, referencedByCount int32, ) *GetConnectorsResponseBodyProperties`

NewGetConnectorsResponseBodyProperties instantiates a new GetConnectorsResponseBodyProperties object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGetConnectorsResponseBodyPropertiesWithDefaults

`func NewGetConnectorsResponseBodyPropertiesWithDefaults() *GetConnectorsResponseBodyProperties`

NewGetConnectorsResponseBodyPropertiesWithDefaults instantiates a new GetConnectorsResponseBodyProperties object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetConnectorTypeId

`func (o *GetConnectorsResponseBodyProperties) GetConnectorTypeId() ConnectorTypes`

GetConnectorTypeId returns the ConnectorTypeId field if non-nil, zero value otherwise.

### GetConnectorTypeIdOk

`func (o *GetConnectorsResponseBodyProperties) GetConnectorTypeIdOk() (*ConnectorTypes, bool)`

GetConnectorTypeIdOk returns a tuple with the ConnectorTypeId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConnectorTypeId

`func (o *GetConnectorsResponseBodyProperties) SetConnectorTypeId(v ConnectorTypes)`

SetConnectorTypeId sets ConnectorTypeId field to given value.


### GetConfig

`func (o *GetConnectorsResponseBodyProperties) GetConfig() map[string]interface{}`

GetConfig returns the Config field if non-nil, zero value otherwise.

### GetConfigOk

`func (o *GetConnectorsResponseBodyProperties) GetConfigOk() (*map[string]interface{}, bool)`

GetConfigOk returns a tuple with the Config field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConfig

`func (o *GetConnectorsResponseBodyProperties) SetConfig(v map[string]interface{})`

SetConfig sets Config field to given value.

### HasConfig

`func (o *GetConnectorsResponseBodyProperties) HasConfig() bool`

HasConfig returns a boolean if a field has been set.

### SetConfigNil

`func (o *GetConnectorsResponseBodyProperties) SetConfigNil(b bool)`

 SetConfigNil sets the value for Config to be an explicit nil

### UnsetConfig
`func (o *GetConnectorsResponseBodyProperties) UnsetConfig()`

UnsetConfig ensures that no value is present for Config, not even an explicit nil
### GetId

`func (o *GetConnectorsResponseBodyProperties) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *GetConnectorsResponseBodyProperties) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *GetConnectorsResponseBodyProperties) SetId(v string)`

SetId sets Id field to given value.


### GetIsDeprecated

`func (o *GetConnectorsResponseBodyProperties) GetIsDeprecated() bool`

GetIsDeprecated returns the IsDeprecated field if non-nil, zero value otherwise.

### GetIsDeprecatedOk

`func (o *GetConnectorsResponseBodyProperties) GetIsDeprecatedOk() (*bool, bool)`

GetIsDeprecatedOk returns a tuple with the IsDeprecated field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIsDeprecated

`func (o *GetConnectorsResponseBodyProperties) SetIsDeprecated(v bool)`

SetIsDeprecated sets IsDeprecated field to given value.


### GetIsMissingSecrets

`func (o *GetConnectorsResponseBodyProperties) GetIsMissingSecrets() bool`

GetIsMissingSecrets returns the IsMissingSecrets field if non-nil, zero value otherwise.

### GetIsMissingSecretsOk

`func (o *GetConnectorsResponseBodyProperties) GetIsMissingSecretsOk() (*bool, bool)`

GetIsMissingSecretsOk returns a tuple with the IsMissingSecrets field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIsMissingSecrets

`func (o *GetConnectorsResponseBodyProperties) SetIsMissingSecrets(v bool)`

SetIsMissingSecrets sets IsMissingSecrets field to given value.

### HasIsMissingSecrets

`func (o *GetConnectorsResponseBodyProperties) HasIsMissingSecrets() bool`

HasIsMissingSecrets returns a boolean if a field has been set.

### GetIsPreconfigured

`func (o *GetConnectorsResponseBodyProperties) GetIsPreconfigured() bool`

GetIsPreconfigured returns the IsPreconfigured field if non-nil, zero value otherwise.

### GetIsPreconfiguredOk

`func (o *GetConnectorsResponseBodyProperties) GetIsPreconfiguredOk() (*bool, bool)`

GetIsPreconfiguredOk returns a tuple with the IsPreconfigured field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIsPreconfigured

`func (o *GetConnectorsResponseBodyProperties) SetIsPreconfigured(v bool)`

SetIsPreconfigured sets IsPreconfigured field to given value.


### GetName

`func (o *GetConnectorsResponseBodyProperties) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *GetConnectorsResponseBodyProperties) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *GetConnectorsResponseBodyProperties) SetName(v string)`

SetName sets Name field to given value.


### GetReferencedByCount

`func (o *GetConnectorsResponseBodyProperties) GetReferencedByCount() int32`

GetReferencedByCount returns the ReferencedByCount field if non-nil, zero value otherwise.

### GetReferencedByCountOk

`func (o *GetConnectorsResponseBodyProperties) GetReferencedByCountOk() (*int32, bool)`

GetReferencedByCountOk returns a tuple with the ReferencedByCount field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReferencedByCount

`func (o *GetConnectorsResponseBodyProperties) SetReferencedByCount(v int32)`

SetReferencedByCount sets ReferencedByCount field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


