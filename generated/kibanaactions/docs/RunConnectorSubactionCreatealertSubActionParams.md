# RunConnectorSubactionCreatealertSubActionParams

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Actions** | Pointer to **[]string** | The custom actions available to the alert. | [optional] 
**Alias** | Pointer to **string** | The unique identifier used for alert deduplication in Opsgenie. | [optional] 
**Description** | Pointer to **string** | A description that provides detailed information about the alert. | [optional] 
**Details** | Pointer to **map[string]interface{}** | The custom properties of the alert. | [optional] 
**Entity** | Pointer to **string** | The domain of the alert. For example, the application or server name. | [optional] 
**Message** | **string** | The alert message. | 
**Note** | Pointer to **string** | Additional information for the alert. | [optional] 
**Priority** | Pointer to **string** | The priority level for the alert. | [optional] 
**Responders** | Pointer to [**[]RunConnectorSubactionCreatealertSubActionParamsRespondersInner**](RunConnectorSubactionCreatealertSubActionParamsRespondersInner.md) | The entities to receive notifications about the alert. If &#x60;type&#x60; is &#x60;user&#x60;, either &#x60;id&#x60; or &#x60;username&#x60; is required. If &#x60;type&#x60; is &#x60;team&#x60;, either &#x60;id&#x60; or &#x60;name&#x60; is required.  | [optional] 
**Source** | Pointer to **string** | The display name for the source of the alert. | [optional] 
**Tags** | Pointer to **[]string** | The tags for the alert. | [optional] 
**User** | Pointer to **string** | The display name for the owner. | [optional] 
**VisibleTo** | Pointer to [**[]RunConnectorSubactionCreatealertSubActionParamsVisibleToInner**](RunConnectorSubactionCreatealertSubActionParamsVisibleToInner.md) | The teams and users that the alert will be visible to without sending a notification. Only one of &#x60;id&#x60;, &#x60;name&#x60;, or &#x60;username&#x60; is required. | [optional] 

## Methods

### NewRunConnectorSubactionCreatealertSubActionParams

`func NewRunConnectorSubactionCreatealertSubActionParams(message string, ) *RunConnectorSubactionCreatealertSubActionParams`

NewRunConnectorSubactionCreatealertSubActionParams instantiates a new RunConnectorSubactionCreatealertSubActionParams object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewRunConnectorSubactionCreatealertSubActionParamsWithDefaults

`func NewRunConnectorSubactionCreatealertSubActionParamsWithDefaults() *RunConnectorSubactionCreatealertSubActionParams`

NewRunConnectorSubactionCreatealertSubActionParamsWithDefaults instantiates a new RunConnectorSubactionCreatealertSubActionParams object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetActions

`func (o *RunConnectorSubactionCreatealertSubActionParams) GetActions() []string`

GetActions returns the Actions field if non-nil, zero value otherwise.

### GetActionsOk

`func (o *RunConnectorSubactionCreatealertSubActionParams) GetActionsOk() (*[]string, bool)`

GetActionsOk returns a tuple with the Actions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetActions

`func (o *RunConnectorSubactionCreatealertSubActionParams) SetActions(v []string)`

SetActions sets Actions field to given value.

### HasActions

`func (o *RunConnectorSubactionCreatealertSubActionParams) HasActions() bool`

HasActions returns a boolean if a field has been set.

### GetAlias

`func (o *RunConnectorSubactionCreatealertSubActionParams) GetAlias() string`

GetAlias returns the Alias field if non-nil, zero value otherwise.

### GetAliasOk

`func (o *RunConnectorSubactionCreatealertSubActionParams) GetAliasOk() (*string, bool)`

GetAliasOk returns a tuple with the Alias field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAlias

`func (o *RunConnectorSubactionCreatealertSubActionParams) SetAlias(v string)`

SetAlias sets Alias field to given value.

### HasAlias

`func (o *RunConnectorSubactionCreatealertSubActionParams) HasAlias() bool`

HasAlias returns a boolean if a field has been set.

### GetDescription

`func (o *RunConnectorSubactionCreatealertSubActionParams) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *RunConnectorSubactionCreatealertSubActionParams) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *RunConnectorSubactionCreatealertSubActionParams) SetDescription(v string)`

SetDescription sets Description field to given value.

### HasDescription

`func (o *RunConnectorSubactionCreatealertSubActionParams) HasDescription() bool`

HasDescription returns a boolean if a field has been set.

### GetDetails

`func (o *RunConnectorSubactionCreatealertSubActionParams) GetDetails() map[string]interface{}`

GetDetails returns the Details field if non-nil, zero value otherwise.

### GetDetailsOk

`func (o *RunConnectorSubactionCreatealertSubActionParams) GetDetailsOk() (*map[string]interface{}, bool)`

GetDetailsOk returns a tuple with the Details field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDetails

`func (o *RunConnectorSubactionCreatealertSubActionParams) SetDetails(v map[string]interface{})`

SetDetails sets Details field to given value.

### HasDetails

`func (o *RunConnectorSubactionCreatealertSubActionParams) HasDetails() bool`

HasDetails returns a boolean if a field has been set.

### GetEntity

`func (o *RunConnectorSubactionCreatealertSubActionParams) GetEntity() string`

GetEntity returns the Entity field if non-nil, zero value otherwise.

### GetEntityOk

`func (o *RunConnectorSubactionCreatealertSubActionParams) GetEntityOk() (*string, bool)`

GetEntityOk returns a tuple with the Entity field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEntity

`func (o *RunConnectorSubactionCreatealertSubActionParams) SetEntity(v string)`

SetEntity sets Entity field to given value.

### HasEntity

`func (o *RunConnectorSubactionCreatealertSubActionParams) HasEntity() bool`

HasEntity returns a boolean if a field has been set.

### GetMessage

`func (o *RunConnectorSubactionCreatealertSubActionParams) GetMessage() string`

GetMessage returns the Message field if non-nil, zero value otherwise.

### GetMessageOk

`func (o *RunConnectorSubactionCreatealertSubActionParams) GetMessageOk() (*string, bool)`

GetMessageOk returns a tuple with the Message field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMessage

`func (o *RunConnectorSubactionCreatealertSubActionParams) SetMessage(v string)`

SetMessage sets Message field to given value.


### GetNote

`func (o *RunConnectorSubactionCreatealertSubActionParams) GetNote() string`

GetNote returns the Note field if non-nil, zero value otherwise.

### GetNoteOk

`func (o *RunConnectorSubactionCreatealertSubActionParams) GetNoteOk() (*string, bool)`

GetNoteOk returns a tuple with the Note field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNote

`func (o *RunConnectorSubactionCreatealertSubActionParams) SetNote(v string)`

SetNote sets Note field to given value.

### HasNote

`func (o *RunConnectorSubactionCreatealertSubActionParams) HasNote() bool`

HasNote returns a boolean if a field has been set.

### GetPriority

`func (o *RunConnectorSubactionCreatealertSubActionParams) GetPriority() string`

GetPriority returns the Priority field if non-nil, zero value otherwise.

### GetPriorityOk

`func (o *RunConnectorSubactionCreatealertSubActionParams) GetPriorityOk() (*string, bool)`

GetPriorityOk returns a tuple with the Priority field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPriority

`func (o *RunConnectorSubactionCreatealertSubActionParams) SetPriority(v string)`

SetPriority sets Priority field to given value.

### HasPriority

`func (o *RunConnectorSubactionCreatealertSubActionParams) HasPriority() bool`

HasPriority returns a boolean if a field has been set.

### GetResponders

`func (o *RunConnectorSubactionCreatealertSubActionParams) GetResponders() []RunConnectorSubactionCreatealertSubActionParamsRespondersInner`

GetResponders returns the Responders field if non-nil, zero value otherwise.

### GetRespondersOk

`func (o *RunConnectorSubactionCreatealertSubActionParams) GetRespondersOk() (*[]RunConnectorSubactionCreatealertSubActionParamsRespondersInner, bool)`

GetRespondersOk returns a tuple with the Responders field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResponders

`func (o *RunConnectorSubactionCreatealertSubActionParams) SetResponders(v []RunConnectorSubactionCreatealertSubActionParamsRespondersInner)`

SetResponders sets Responders field to given value.

### HasResponders

`func (o *RunConnectorSubactionCreatealertSubActionParams) HasResponders() bool`

HasResponders returns a boolean if a field has been set.

### GetSource

`func (o *RunConnectorSubactionCreatealertSubActionParams) GetSource() string`

GetSource returns the Source field if non-nil, zero value otherwise.

### GetSourceOk

`func (o *RunConnectorSubactionCreatealertSubActionParams) GetSourceOk() (*string, bool)`

GetSourceOk returns a tuple with the Source field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSource

`func (o *RunConnectorSubactionCreatealertSubActionParams) SetSource(v string)`

SetSource sets Source field to given value.

### HasSource

`func (o *RunConnectorSubactionCreatealertSubActionParams) HasSource() bool`

HasSource returns a boolean if a field has been set.

### GetTags

`func (o *RunConnectorSubactionCreatealertSubActionParams) GetTags() []string`

GetTags returns the Tags field if non-nil, zero value otherwise.

### GetTagsOk

`func (o *RunConnectorSubactionCreatealertSubActionParams) GetTagsOk() (*[]string, bool)`

GetTagsOk returns a tuple with the Tags field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTags

`func (o *RunConnectorSubactionCreatealertSubActionParams) SetTags(v []string)`

SetTags sets Tags field to given value.

### HasTags

`func (o *RunConnectorSubactionCreatealertSubActionParams) HasTags() bool`

HasTags returns a boolean if a field has been set.

### GetUser

`func (o *RunConnectorSubactionCreatealertSubActionParams) GetUser() string`

GetUser returns the User field if non-nil, zero value otherwise.

### GetUserOk

`func (o *RunConnectorSubactionCreatealertSubActionParams) GetUserOk() (*string, bool)`

GetUserOk returns a tuple with the User field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUser

`func (o *RunConnectorSubactionCreatealertSubActionParams) SetUser(v string)`

SetUser sets User field to given value.

### HasUser

`func (o *RunConnectorSubactionCreatealertSubActionParams) HasUser() bool`

HasUser returns a boolean if a field has been set.

### GetVisibleTo

`func (o *RunConnectorSubactionCreatealertSubActionParams) GetVisibleTo() []RunConnectorSubactionCreatealertSubActionParamsVisibleToInner`

GetVisibleTo returns the VisibleTo field if non-nil, zero value otherwise.

### GetVisibleToOk

`func (o *RunConnectorSubactionCreatealertSubActionParams) GetVisibleToOk() (*[]RunConnectorSubactionCreatealertSubActionParamsVisibleToInner, bool)`

GetVisibleToOk returns a tuple with the VisibleTo field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetVisibleTo

`func (o *RunConnectorSubactionCreatealertSubActionParams) SetVisibleTo(v []RunConnectorSubactionCreatealertSubActionParamsVisibleToInner)`

SetVisibleTo sets VisibleTo field to given value.

### HasVisibleTo

`func (o *RunConnectorSubactionCreatealertSubActionParams) HasVisibleTo() bool`

HasVisibleTo returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


