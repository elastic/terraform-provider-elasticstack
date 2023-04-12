# RunConnectorSubactionClosealertSubActionParams

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Alias** | **string** | The unique identifier used for alert deduplication in Opsgenie. The alias must match the value used when creating the alert. | 
**Note** | Pointer to **string** | Additional information for the alert. | [optional] 
**Source** | Pointer to **string** | The display name for the source of the alert. | [optional] 
**User** | Pointer to **string** | The display name for the owner. | [optional] 

## Methods

### NewRunConnectorSubactionClosealertSubActionParams

`func NewRunConnectorSubactionClosealertSubActionParams(alias string, ) *RunConnectorSubactionClosealertSubActionParams`

NewRunConnectorSubactionClosealertSubActionParams instantiates a new RunConnectorSubactionClosealertSubActionParams object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewRunConnectorSubactionClosealertSubActionParamsWithDefaults

`func NewRunConnectorSubactionClosealertSubActionParamsWithDefaults() *RunConnectorSubactionClosealertSubActionParams`

NewRunConnectorSubactionClosealertSubActionParamsWithDefaults instantiates a new RunConnectorSubactionClosealertSubActionParams object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAlias

`func (o *RunConnectorSubactionClosealertSubActionParams) GetAlias() string`

GetAlias returns the Alias field if non-nil, zero value otherwise.

### GetAliasOk

`func (o *RunConnectorSubactionClosealertSubActionParams) GetAliasOk() (*string, bool)`

GetAliasOk returns a tuple with the Alias field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAlias

`func (o *RunConnectorSubactionClosealertSubActionParams) SetAlias(v string)`

SetAlias sets Alias field to given value.


### GetNote

`func (o *RunConnectorSubactionClosealertSubActionParams) GetNote() string`

GetNote returns the Note field if non-nil, zero value otherwise.

### GetNoteOk

`func (o *RunConnectorSubactionClosealertSubActionParams) GetNoteOk() (*string, bool)`

GetNoteOk returns a tuple with the Note field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNote

`func (o *RunConnectorSubactionClosealertSubActionParams) SetNote(v string)`

SetNote sets Note field to given value.

### HasNote

`func (o *RunConnectorSubactionClosealertSubActionParams) HasNote() bool`

HasNote returns a boolean if a field has been set.

### GetSource

`func (o *RunConnectorSubactionClosealertSubActionParams) GetSource() string`

GetSource returns the Source field if non-nil, zero value otherwise.

### GetSourceOk

`func (o *RunConnectorSubactionClosealertSubActionParams) GetSourceOk() (*string, bool)`

GetSourceOk returns a tuple with the Source field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSource

`func (o *RunConnectorSubactionClosealertSubActionParams) SetSource(v string)`

SetSource sets Source field to given value.

### HasSource

`func (o *RunConnectorSubactionClosealertSubActionParams) HasSource() bool`

HasSource returns a boolean if a field has been set.

### GetUser

`func (o *RunConnectorSubactionClosealertSubActionParams) GetUser() string`

GetUser returns the User field if non-nil, zero value otherwise.

### GetUserOk

`func (o *RunConnectorSubactionClosealertSubActionParams) GetUserOk() (*string, bool)`

GetUserOk returns a tuple with the User field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUser

`func (o *RunConnectorSubactionClosealertSubActionParams) SetUser(v string)`

SetUser sets User field to given value.

### HasUser

`func (o *RunConnectorSubactionClosealertSubActionParams) HasUser() bool`

HasUser returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


