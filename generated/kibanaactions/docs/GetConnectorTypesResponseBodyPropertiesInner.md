# GetConnectorTypesResponseBodyPropertiesInner

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Enabled** | Pointer to **bool** | Indicates whether the connector type is enabled in Kibana. | [optional] 
**EnabledInConfig** | Pointer to **bool** | Indicates whether the connector type is enabled in the Kibana &#x60;.yml&#x60; file. | [optional] 
**EnabledInLicense** | Pointer to **bool** | Indicates whether the connector is enabled in the license. | [optional] 
**Id** | Pointer to [**ConnectorTypes**](ConnectorTypes.md) |  | [optional] 
**MinimumLicenseRequired** | Pointer to **string** | The license that is required to use the connector type. | [optional] 
**Name** | Pointer to **string** | The name of the connector type. | [optional] 
**SupportedFeatureIds** | Pointer to [**[]Features**](Features.md) | The Kibana features that are supported by the connector type. | [optional] 

## Methods

### NewGetConnectorTypesResponseBodyPropertiesInner

`func NewGetConnectorTypesResponseBodyPropertiesInner() *GetConnectorTypesResponseBodyPropertiesInner`

NewGetConnectorTypesResponseBodyPropertiesInner instantiates a new GetConnectorTypesResponseBodyPropertiesInner object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGetConnectorTypesResponseBodyPropertiesInnerWithDefaults

`func NewGetConnectorTypesResponseBodyPropertiesInnerWithDefaults() *GetConnectorTypesResponseBodyPropertiesInner`

NewGetConnectorTypesResponseBodyPropertiesInnerWithDefaults instantiates a new GetConnectorTypesResponseBodyPropertiesInner object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetEnabled

`func (o *GetConnectorTypesResponseBodyPropertiesInner) GetEnabled() bool`

GetEnabled returns the Enabled field if non-nil, zero value otherwise.

### GetEnabledOk

`func (o *GetConnectorTypesResponseBodyPropertiesInner) GetEnabledOk() (*bool, bool)`

GetEnabledOk returns a tuple with the Enabled field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnabled

`func (o *GetConnectorTypesResponseBodyPropertiesInner) SetEnabled(v bool)`

SetEnabled sets Enabled field to given value.

### HasEnabled

`func (o *GetConnectorTypesResponseBodyPropertiesInner) HasEnabled() bool`

HasEnabled returns a boolean if a field has been set.

### GetEnabledInConfig

`func (o *GetConnectorTypesResponseBodyPropertiesInner) GetEnabledInConfig() bool`

GetEnabledInConfig returns the EnabledInConfig field if non-nil, zero value otherwise.

### GetEnabledInConfigOk

`func (o *GetConnectorTypesResponseBodyPropertiesInner) GetEnabledInConfigOk() (*bool, bool)`

GetEnabledInConfigOk returns a tuple with the EnabledInConfig field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnabledInConfig

`func (o *GetConnectorTypesResponseBodyPropertiesInner) SetEnabledInConfig(v bool)`

SetEnabledInConfig sets EnabledInConfig field to given value.

### HasEnabledInConfig

`func (o *GetConnectorTypesResponseBodyPropertiesInner) HasEnabledInConfig() bool`

HasEnabledInConfig returns a boolean if a field has been set.

### GetEnabledInLicense

`func (o *GetConnectorTypesResponseBodyPropertiesInner) GetEnabledInLicense() bool`

GetEnabledInLicense returns the EnabledInLicense field if non-nil, zero value otherwise.

### GetEnabledInLicenseOk

`func (o *GetConnectorTypesResponseBodyPropertiesInner) GetEnabledInLicenseOk() (*bool, bool)`

GetEnabledInLicenseOk returns a tuple with the EnabledInLicense field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnabledInLicense

`func (o *GetConnectorTypesResponseBodyPropertiesInner) SetEnabledInLicense(v bool)`

SetEnabledInLicense sets EnabledInLicense field to given value.

### HasEnabledInLicense

`func (o *GetConnectorTypesResponseBodyPropertiesInner) HasEnabledInLicense() bool`

HasEnabledInLicense returns a boolean if a field has been set.

### GetId

`func (o *GetConnectorTypesResponseBodyPropertiesInner) GetId() ConnectorTypes`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *GetConnectorTypesResponseBodyPropertiesInner) GetIdOk() (*ConnectorTypes, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *GetConnectorTypesResponseBodyPropertiesInner) SetId(v ConnectorTypes)`

SetId sets Id field to given value.

### HasId

`func (o *GetConnectorTypesResponseBodyPropertiesInner) HasId() bool`

HasId returns a boolean if a field has been set.

### GetMinimumLicenseRequired

`func (o *GetConnectorTypesResponseBodyPropertiesInner) GetMinimumLicenseRequired() string`

GetMinimumLicenseRequired returns the MinimumLicenseRequired field if non-nil, zero value otherwise.

### GetMinimumLicenseRequiredOk

`func (o *GetConnectorTypesResponseBodyPropertiesInner) GetMinimumLicenseRequiredOk() (*string, bool)`

GetMinimumLicenseRequiredOk returns a tuple with the MinimumLicenseRequired field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMinimumLicenseRequired

`func (o *GetConnectorTypesResponseBodyPropertiesInner) SetMinimumLicenseRequired(v string)`

SetMinimumLicenseRequired sets MinimumLicenseRequired field to given value.

### HasMinimumLicenseRequired

`func (o *GetConnectorTypesResponseBodyPropertiesInner) HasMinimumLicenseRequired() bool`

HasMinimumLicenseRequired returns a boolean if a field has been set.

### GetName

`func (o *GetConnectorTypesResponseBodyPropertiesInner) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *GetConnectorTypesResponseBodyPropertiesInner) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *GetConnectorTypesResponseBodyPropertiesInner) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *GetConnectorTypesResponseBodyPropertiesInner) HasName() bool`

HasName returns a boolean if a field has been set.

### GetSupportedFeatureIds

`func (o *GetConnectorTypesResponseBodyPropertiesInner) GetSupportedFeatureIds() []Features`

GetSupportedFeatureIds returns the SupportedFeatureIds field if non-nil, zero value otherwise.

### GetSupportedFeatureIdsOk

`func (o *GetConnectorTypesResponseBodyPropertiesInner) GetSupportedFeatureIdsOk() (*[]Features, bool)`

GetSupportedFeatureIdsOk returns a tuple with the SupportedFeatureIds field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSupportedFeatureIds

`func (o *GetConnectorTypesResponseBodyPropertiesInner) SetSupportedFeatureIds(v []Features)`

SetSupportedFeatureIds sets SupportedFeatureIds field to given value.

### HasSupportedFeatureIds

`func (o *GetConnectorTypesResponseBodyPropertiesInner) HasSupportedFeatureIds() bool`

HasSupportedFeatureIds returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


