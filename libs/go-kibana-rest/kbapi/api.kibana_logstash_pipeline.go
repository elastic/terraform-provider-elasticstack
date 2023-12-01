package kbapi

import (
	"encoding/json"
	"fmt"

	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
)

const (
	basePathKibanaLogstashPipeline = "/api/logstash/pipeline" // Base URL to access on Kibana Logstash pipeline
)

// LogstashPipeline is the Logstash pipeline object
type LogstashPipeline struct {
	ID          string                 `json:"id"`
	Description string                 `json:"description,omitempty"`
	Pipeline    string                 `json:"pipeline,omitempty"`
	Settings    map[string]interface{} `json:"settings,omitempty"`
	Username    string                 `json:"username,omitempty"`
}

type LogstashPipelineRequest struct {
	Description string                 `json:"description,omitempty"`
	Pipeline    string                 `json:"pipeline,omitempty"`
	Settings    map[string]interface{} `json:"settings,omitempty"`
	Username    string                 `json:"username,omitempty"`
}

// LogstashPipelinesList is the logstash pipeline list result when get the list
type LogstashPipelinesList struct {
	Pipelines LogstashPipelines `json:"pipelines"`
}

// LogstashPipelines is list of Logstash pipeline object
type LogstashPipelines []LogstashPipeline

// KibanaLogstashPipelineCreateOrUpdate permit to create or update logstash pipeline
type KibanaLogstashPipelineCreateOrUpdate func(logstashPipeline *LogstashPipeline) (*LogstashPipeline, error)

// KibanaLogstashPipelineGet permit to get the logstash pipeline
type KibanaLogstashPipelineGet func(id string) (*LogstashPipeline, error)

// KibanaLogstashPipelineList permit to get all the logstash pipeline
type KibanaLogstashPipelineList func() (LogstashPipelines, error)

// KibanaLogstashPipelineDelete permit to delete the logstash pipeline
type KibanaLogstashPipelineDelete func(id string) error

// String permit to return LogstashPipeline object as JSON string
func (o *LogstashPipeline) String() string {
	json, _ := json.Marshal(o)
	return string(json)
}

// newKibanaLogstashPipelineGetFunc permit to get the kibana role with it name
func newKibanaLogstashPipelineGetFunc(c *resty.Client) KibanaLogstashPipelineGet {
	return func(id string) (*LogstashPipeline, error) {

		if id == "" {
			return nil, NewAPIError(600, "You must provide logstash pipline ID")
		}
		log.Debug("ID: ", id)

		path := fmt.Sprintf("%s/%s", basePathKibanaLogstashPipeline, id)
		resp, err := c.R().Get(path)
		if err != nil {
			return nil, err
		}
		log.Debug("Response: ", resp)
		if resp.StatusCode() >= 300 {
			if resp.StatusCode() == 404 {
				return nil, nil
			}
			return nil, NewAPIError(resp.StatusCode(), resp.Status())
		}
		logstashPipeline := &LogstashPipeline{}
		err = json.Unmarshal(resp.Body(), logstashPipeline)
		if err != nil {
			return nil, err
		}
		log.Debug("LogstashPipeline: ", logstashPipeline)

		return logstashPipeline, nil
	}
}

// newKibanaLogstashPipelineListFunc permit to get all kibana role
func newKibanaLogstashPipelineListFunc(c *resty.Client) KibanaLogstashPipelineList {
	return func() (LogstashPipelines, error) {

		path := fmt.Sprintf("%ss", basePathKibanaLogstashPipeline)
		resp, err := c.R().Get(path)
		if err != nil {
			return nil, err
		}
		log.Debug("Response: ", resp)
		if resp.StatusCode() >= 300 {
			return nil, NewAPIError(resp.StatusCode(), resp.Status())
		}
		logstashPipelinesList := &LogstashPipelinesList{}
		err = json.Unmarshal(resp.Body(), logstashPipelinesList)
		if err != nil {
			return nil, err
		}
		log.Debug("LogstashPipelines: ", logstashPipelinesList)

		return logstashPipelinesList.Pipelines, nil
	}

}

// newKibanaPipelineCreateOrUpdateFunc permit to create or update logstash pipeline
func newKibanaLogstashPipelineCreateOrUpdateFunc(c *resty.Client) KibanaLogstashPipelineCreateOrUpdate {
	return func(logstashPipeline *LogstashPipeline) (*LogstashPipeline, error) {

		if logstashPipeline == nil {
			return nil, NewAPIError(600, "You must provide the logstash pipeline object")
		}

		log.Debug("LogstashPipeline: ", logstashPipeline)

		logstashPipelineRequest := &LogstashPipelineRequest{
			Description: logstashPipeline.Description,
			Pipeline:    logstashPipeline.Pipeline,
			Settings:    logstashPipeline.Settings,
		}

		jsonData, err := json.Marshal(logstashPipelineRequest)
		if err != nil {
			return nil, err
		}

		path := fmt.Sprintf("%s/%s", basePathKibanaLogstashPipeline, logstashPipeline.ID)
		resp, err := c.R().SetBody(jsonData).Put(path)
		if err != nil {
			return nil, err
		}

		log.Debug("Response: ", resp)
		if resp.StatusCode() >= 300 {
			return nil, NewAPIError(resp.StatusCode(), resp.Status())
		}

		// Retrive the object to return it
		logstashPipeline, err = newKibanaLogstashPipelineGetFunc(c)(logstashPipeline.ID)
		if err != nil {
			return nil, err
		}
		if logstashPipeline == nil {
			return nil, NewAPIError(404, "Logstash pipeline %s not found", logstashPipeline.ID)
		}

		log.Debug("logstashPipeline: ", logstashPipeline)

		return logstashPipeline, nil
	}
}

// newKibanaLogstashPipelineDeleteFunc permit to delete logstash pipeline with it ID
func newKibanaLogstashPipelineDeleteFunc(c *resty.Client) KibanaLogstashPipelineDelete {
	return func(id string) error {

		if id == "" {
			return NewAPIError(600, "You must provide logstash pipeline ID")
		}
		log.Debug("ID: ", id)

		path := fmt.Sprintf("%s/%s", basePathKibanaLogstashPipeline, id)
		resp, err := c.R().Delete(path)
		if err != nil {
			return err
		}
		log.Debug("Response: ", resp)
		if resp.StatusCode() >= 300 {
			return NewAPIError(resp.StatusCode(), resp.Status())
		}

		return nil
	}
}
