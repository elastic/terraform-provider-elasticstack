# LegacyGetConnectorTypesResponseBodyPropertiesInner

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Enabled** | Pointer to **bool** | Indicates whether the connector type is enabled in Kibana. | [optional] 
**EnabledInConfig** | Pointer to **bool** | Indicates whether the connector type is enabled in the Kibana &#x60;.yml&#x60; file. | [optional] 
**EnabledInLicense** | Pointer to **bool** | Indicates whether the connector is enabled in the license. | [optional] 
**Id** | Pointer to **string** | The unique identifier for the connector type. | [optional] 
**MinimumLicenseRequired** | Pointer to **string** | The license that is required to use the connector type. | [optional] 
**Name** | Pointer to **string** | The name of the connector type. | [optional] 

## Methods

### NewLegacyGetConnectorTypesResponseBodyPropertiesInner

`func NewLegacyGetConnectorTypesResponseBodyPropertiesInner() *LegacyGetConnectorTypesResponseBodyPropertiesInner`

NewLegacyGetConnectorTypesResponseBodyPropertiesInner instantiates a new LegacyGetConnectorTypesResponseBodyPropertiesInner object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewLegacyGetConnectorTypesResponseBodyPropertiesInnerWithDefaults

`func NewLegacyGetConnectorTypesResponseBodyPropertiesInnerWithDefaults() *LegacyGetConnectorTypesResponseBodyPropertiesInner`

NewLegacyGetConnectorTypesResponseBodyPropertiesInnerWithDefaults instantiates a new LegacyGetConnectorTypesResponseBodyPropertiesInner object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetEnabled

`func (o *LegacyGetConnectorTypesResponseBodyPropertiesInner) GetEnabled() bool`

GetEnabled returns the Enabled field if non-nil, zero value otherwise.

### GetEnabledOk

`func (o *LegacyGetConnectorTypesResponseBodyPropertiesInner) GetEnabledOk() (*bool, bool)`

GetEnabledOk returns a tuple with the Enabled field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnabled

`func (o *LegacyGetConnectorTypesResponseBodyPropertiesInner) SetEnabled(v bool)`

SetEnabled sets Enabled field to given value.

### HasEnabled

`func (o *LegacyGetConnectorTypesResponseBodyPropertiesInner) HasEnabled() bool`

HasEnabled returns a boolean if a field has been set.

### GetEnabledInConfig

`func (o *LegacyGetConnectorTypesResponseBodyPropertiesInner) GetEnabledInConfig() bool`

GetEnabledInConfig returns the EnabledInConfig field if non-nil, zero value otherwise.

### GetEnabledInConfigOk

`func (o *LegacyGetConnectorTypesResponseBodyPropertiesInner) GetEnabledInConfigOk() (*bool, bool)`

GetEnabledInConfigOk returns a tuple with the EnabledInConfig field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnabledInConfig

`func (o *LegacyGetConnectorTypesResponseBodyPropertiesInner) SetEnabledInConfig(v bool)`

SetEnabledInConfig sets EnabledInConfig field to given value.

### HasEnabledInConfig

`func (o *LegacyGetConnectorTypesResponseBodyPropertiesInner) HasEnabledInConfig() bool`

HasEnabledInConfig returns a boolean if a field has been set.

### GetEnabledInLicense

`func (o *LegacyGetConnectorTypesResponseBodyPropertiesInner) GetEnabledInLicense() bool`

GetEnabledInLicense returns the EnabledInLicense field if non-nil, zero value otherwise.

### GetEnabledInLicenseOk

`func (o *LegacyGetConnectorTypesResponseBodyPropertiesInner) GetEnabledInLicenseOk() (*bool, bool)`

GetEnabledInLicenseOk returns a tuple with the EnabledInLicense field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnabledInLicense

`func (o *LegacyGetConnectorTypesResponseBodyPropertiesInner) SetEnabledInLicense(v bool)`

SetEnabledInLicense sets EnabledInLicense field to given value.

### HasEnabledInLicense

`func (o *LegacyGetConnectorTypesResponseBodyPropertiesInner) HasEnabledInLicense() bool`

HasEnabledInLicense returns a boolean if a field has been set.

### GetId

`func (o *LegacyGetConnectorTypesResponseBodyPropertiesInner) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *LegacyGetConnectorTypesResponseBodyPropertiesInner) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *LegacyGetConnectorTypesResponseBodyPropertiesInner) SetId(v string)`

SetId sets Id field to given value.

### HasId

`func (o *LegacyGetConnectorTypesResponseBodyPropertiesInner) HasId() bool`

HasId returns a boolean if a field has been set.

### GetMinimumLicenseRequired

`func (o *LegacyGetConnectorTypesResponseBodyPropertiesInner) GetMinimumLicenseRequired() string`

GetMinimumLicenseRequired returns the MinimumLicenseRequired field if non-nil, zero value otherwise.

### GetMinimumLicenseRequiredOk

`func (o *LegacyGetConnectorTypesResponseBodyPropertiesInner) GetMinimumLicenseRequiredOk() (*string, bool)`

GetMinimumLicenseRequiredOk returns a tuple with the MinimumLicenseRequired field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMinimumLicenseRequired

`func (o *LegacyGetConnectorTypesResponseBodyPropertiesInner) SetMinimumLicenseRequired(v string)`

SetMinimumLicenseRequired sets MinimumLicenseRequired field to given value.

### HasMinimumLicenseRequired

`func (o *LegacyGetConnectorTypesResponseBodyPropertiesInner) HasMinimumLicenseRequired() bool`

HasMinimumLicenseRequired returns a boolean if a field has been set.

### GetName

`func (o *LegacyGetConnectorTypesResponseBodyPropertiesInner) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *LegacyGetConnectorTypesResponseBodyPropertiesInner) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *LegacyGetConnectorTypesResponseBodyPropertiesInner) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *LegacyGetConnectorTypesResponseBodyPropertiesInner) HasName() bool`

HasName returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


