# RunConnectorSubactionAddeventSubActionParams

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**AdditionalInfo** | Pointer to **string** | Additional information about the event. | [optional] 
**Description** | Pointer to **string** | The details about the event. | [optional] 
**EventClass** | Pointer to **string** | A specific instance of the source. | [optional] 
**MessageKey** | Pointer to **string** | All actions sharing this key are associated with the same ServiceNow alert. The default value is &#x60;&lt;rule ID&gt;:&lt;alert instance ID&gt;&#x60;. | [optional] 
**MetricName** | Pointer to **string** | The name of the metric. | [optional] 
**Node** | Pointer to **string** | The host that the event was triggered for. | [optional] 
**Resource** | Pointer to **string** | The name of the resource. | [optional] 
**Severity** | Pointer to **string** | The severity of the event. | [optional] 
**Source** | Pointer to **string** | The name of the event source type. | [optional] 
**TimeOfEvent** | Pointer to **string** | The time of the event. | [optional] 
**Type** | Pointer to **string** | The type of event. | [optional] 

## Methods

### NewRunConnectorSubactionAddeventSubActionParams

`func NewRunConnectorSubactionAddeventSubActionParams() *RunConnectorSubactionAddeventSubActionParams`

NewRunConnectorSubactionAddeventSubActionParams instantiates a new RunConnectorSubactionAddeventSubActionParams object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewRunConnectorSubactionAddeventSubActionParamsWithDefaults

`func NewRunConnectorSubactionAddeventSubActionParamsWithDefaults() *RunConnectorSubactionAddeventSubActionParams`

NewRunConnectorSubactionAddeventSubActionParamsWithDefaults instantiates a new RunConnectorSubactionAddeventSubActionParams object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAdditionalInfo

`func (o *RunConnectorSubactionAddeventSubActionParams) GetAdditionalInfo() string`

GetAdditionalInfo returns the AdditionalInfo field if non-nil, zero value otherwise.

### GetAdditionalInfoOk

`func (o *RunConnectorSubactionAddeventSubActionParams) GetAdditionalInfoOk() (*string, bool)`

GetAdditionalInfoOk returns a tuple with the AdditionalInfo field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAdditionalInfo

`func (o *RunConnectorSubactionAddeventSubActionParams) SetAdditionalInfo(v string)`

SetAdditionalInfo sets AdditionalInfo field to given value.

### HasAdditionalInfo

`func (o *RunConnectorSubactionAddeventSubActionParams) HasAdditionalInfo() bool`

HasAdditionalInfo returns a boolean if a field has been set.

### GetDescription

`func (o *RunConnectorSubactionAddeventSubActionParams) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *RunConnectorSubactionAddeventSubActionParams) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *RunConnectorSubactionAddeventSubActionParams) SetDescription(v string)`

SetDescription sets Description field to given value.

### HasDescription

`func (o *RunConnectorSubactionAddeventSubActionParams) HasDescription() bool`

HasDescription returns a boolean if a field has been set.

### GetEventClass

`func (o *RunConnectorSubactionAddeventSubActionParams) GetEventClass() string`

GetEventClass returns the EventClass field if non-nil, zero value otherwise.

### GetEventClassOk

`func (o *RunConnectorSubactionAddeventSubActionParams) GetEventClassOk() (*string, bool)`

GetEventClassOk returns a tuple with the EventClass field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEventClass

`func (o *RunConnectorSubactionAddeventSubActionParams) SetEventClass(v string)`

SetEventClass sets EventClass field to given value.

### HasEventClass

`func (o *RunConnectorSubactionAddeventSubActionParams) HasEventClass() bool`

HasEventClass returns a boolean if a field has been set.

### GetMessageKey

`func (o *RunConnectorSubactionAddeventSubActionParams) GetMessageKey() string`

GetMessageKey returns the MessageKey field if non-nil, zero value otherwise.

### GetMessageKeyOk

`func (o *RunConnectorSubactionAddeventSubActionParams) GetMessageKeyOk() (*string, bool)`

GetMessageKeyOk returns a tuple with the MessageKey field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMessageKey

`func (o *RunConnectorSubactionAddeventSubActionParams) SetMessageKey(v string)`

SetMessageKey sets MessageKey field to given value.

### HasMessageKey

`func (o *RunConnectorSubactionAddeventSubActionParams) HasMessageKey() bool`

HasMessageKey returns a boolean if a field has been set.

### GetMetricName

`func (o *RunConnectorSubactionAddeventSubActionParams) GetMetricName() string`

GetMetricName returns the MetricName field if non-nil, zero value otherwise.

### GetMetricNameOk

`func (o *RunConnectorSubactionAddeventSubActionParams) GetMetricNameOk() (*string, bool)`

GetMetricNameOk returns a tuple with the MetricName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMetricName

`func (o *RunConnectorSubactionAddeventSubActionParams) SetMetricName(v string)`

SetMetricName sets MetricName field to given value.

### HasMetricName

`func (o *RunConnectorSubactionAddeventSubActionParams) HasMetricName() bool`

HasMetricName returns a boolean if a field has been set.

### GetNode

`func (o *RunConnectorSubactionAddeventSubActionParams) GetNode() string`

GetNode returns the Node field if non-nil, zero value otherwise.

### GetNodeOk

`func (o *RunConnectorSubactionAddeventSubActionParams) GetNodeOk() (*string, bool)`

GetNodeOk returns a tuple with the Node field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNode

`func (o *RunConnectorSubactionAddeventSubActionParams) SetNode(v string)`

SetNode sets Node field to given value.

### HasNode

`func (o *RunConnectorSubactionAddeventSubActionParams) HasNode() bool`

HasNode returns a boolean if a field has been set.

### GetResource

`func (o *RunConnectorSubactionAddeventSubActionParams) GetResource() string`

GetResource returns the Resource field if non-nil, zero value otherwise.

### GetResourceOk

`func (o *RunConnectorSubactionAddeventSubActionParams) GetResourceOk() (*string, bool)`

GetResourceOk returns a tuple with the Resource field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResource

`func (o *RunConnectorSubactionAddeventSubActionParams) SetResource(v string)`

SetResource sets Resource field to given value.

### HasResource

`func (o *RunConnectorSubactionAddeventSubActionParams) HasResource() bool`

HasResource returns a boolean if a field has been set.

### GetSeverity

`func (o *RunConnectorSubactionAddeventSubActionParams) GetSeverity() string`

GetSeverity returns the Severity field if non-nil, zero value otherwise.

### GetSeverityOk

`func (o *RunConnectorSubactionAddeventSubActionParams) GetSeverityOk() (*string, bool)`

GetSeverityOk returns a tuple with the Severity field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSeverity

`func (o *RunConnectorSubactionAddeventSubActionParams) SetSeverity(v string)`

SetSeverity sets Severity field to given value.

### HasSeverity

`func (o *RunConnectorSubactionAddeventSubActionParams) HasSeverity() bool`

HasSeverity returns a boolean if a field has been set.

### GetSource

`func (o *RunConnectorSubactionAddeventSubActionParams) GetSource() string`

GetSource returns the Source field if non-nil, zero value otherwise.

### GetSourceOk

`func (o *RunConnectorSubactionAddeventSubActionParams) GetSourceOk() (*string, bool)`

GetSourceOk returns a tuple with the Source field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSource

`func (o *RunConnectorSubactionAddeventSubActionParams) SetSource(v string)`

SetSource sets Source field to given value.

### HasSource

`func (o *RunConnectorSubactionAddeventSubActionParams) HasSource() bool`

HasSource returns a boolean if a field has been set.

### GetTimeOfEvent

`func (o *RunConnectorSubactionAddeventSubActionParams) GetTimeOfEvent() string`

GetTimeOfEvent returns the TimeOfEvent field if non-nil, zero value otherwise.

### GetTimeOfEventOk

`func (o *RunConnectorSubactionAddeventSubActionParams) GetTimeOfEventOk() (*string, bool)`

GetTimeOfEventOk returns a tuple with the TimeOfEvent field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTimeOfEvent

`func (o *RunConnectorSubactionAddeventSubActionParams) SetTimeOfEvent(v string)`

SetTimeOfEvent sets TimeOfEvent field to given value.

### HasTimeOfEvent

`func (o *RunConnectorSubactionAddeventSubActionParams) HasTimeOfEvent() bool`

HasTimeOfEvent returns a boolean if a field has been set.

### GetType

`func (o *RunConnectorSubactionAddeventSubActionParams) GetType() string`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *RunConnectorSubactionAddeventSubActionParams) GetTypeOk() (*string, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *RunConnectorSubactionAddeventSubActionParams) SetType(v string)`

SetType sets Type field to given value.

### HasType

`func (o *RunConnectorSubactionAddeventSubActionParams) HasType() bool`

HasType returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


