# RunConnectorSubactionPushtoserviceSubActionParams

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Comments** | Pointer to [**[]RunConnectorSubactionPushtoserviceSubActionParamsCommentsInner**](RunConnectorSubactionPushtoserviceSubActionParamsCommentsInner.md) | Additional information that is sent to Jira, ServiceNow ITSM, ServiceNow SecOps, or Swimlane. | [optional] 
**Incident** | Pointer to [**RunConnectorSubactionPushtoserviceSubActionParamsIncident**](RunConnectorSubactionPushtoserviceSubActionParamsIncident.md) |  | [optional] 

## Methods

### NewRunConnectorSubactionPushtoserviceSubActionParams

`func NewRunConnectorSubactionPushtoserviceSubActionParams() *RunConnectorSubactionPushtoserviceSubActionParams`

NewRunConnectorSubactionPushtoserviceSubActionParams instantiates a new RunConnectorSubactionPushtoserviceSubActionParams object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewRunConnectorSubactionPushtoserviceSubActionParamsWithDefaults

`func NewRunConnectorSubactionPushtoserviceSubActionParamsWithDefaults() *RunConnectorSubactionPushtoserviceSubActionParams`

NewRunConnectorSubactionPushtoserviceSubActionParamsWithDefaults instantiates a new RunConnectorSubactionPushtoserviceSubActionParams object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetComments

`func (o *RunConnectorSubactionPushtoserviceSubActionParams) GetComments() []RunConnectorSubactionPushtoserviceSubActionParamsCommentsInner`

GetComments returns the Comments field if non-nil, zero value otherwise.

### GetCommentsOk

`func (o *RunConnectorSubactionPushtoserviceSubActionParams) GetCommentsOk() (*[]RunConnectorSubactionPushtoserviceSubActionParamsCommentsInner, bool)`

GetCommentsOk returns a tuple with the Comments field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetComments

`func (o *RunConnectorSubactionPushtoserviceSubActionParams) SetComments(v []RunConnectorSubactionPushtoserviceSubActionParamsCommentsInner)`

SetComments sets Comments field to given value.

### HasComments

`func (o *RunConnectorSubactionPushtoserviceSubActionParams) HasComments() bool`

HasComments returns a boolean if a field has been set.

### GetIncident

`func (o *RunConnectorSubactionPushtoserviceSubActionParams) GetIncident() RunConnectorSubactionPushtoserviceSubActionParamsIncident`

GetIncident returns the Incident field if non-nil, zero value otherwise.

### GetIncidentOk

`func (o *RunConnectorSubactionPushtoserviceSubActionParams) GetIncidentOk() (*RunConnectorSubactionPushtoserviceSubActionParamsIncident, bool)`

GetIncidentOk returns a tuple with the Incident field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIncident

`func (o *RunConnectorSubactionPushtoserviceSubActionParams) SetIncident(v RunConnectorSubactionPushtoserviceSubActionParamsIncident)`

SetIncident sets Incident field to given value.

### HasIncident

`func (o *RunConnectorSubactionPushtoserviceSubActionParams) HasIncident() bool`

HasIncident returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


