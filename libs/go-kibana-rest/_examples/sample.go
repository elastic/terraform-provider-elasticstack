package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/disaster37/go-kibana-rest/v8"
	"github.com/disaster37/go-kibana-rest/v8/kbapi"
)

func main() {

	cfg := kibana.Config{
		Address:          "http://127.0.0.1:5601",
		Username:         "elastic",
		Password:         "changeme",
		DisableVerifySSL: true,
	}

	client, err := kibana.NewClient(cfg)

	if err != nil {
		log.Fatalf("Error creating the client: %s", err)
	}

	status, err := client.API.KibanaStatus.Get()
	if err != nil {
		log.Fatalf("Error getting response: %s", err)
	}
	log.Println(status)

	// Shorten long URL
	shortenURL := &kbapi.ShortenURL{
		URL: "/app/kibana#/dashboard?_g=()&_a=(description:'',filters:!(),fullScreenMode:!f,options:(hidePanelTitles:!f,useMargins:!t),panels:!((embeddableConfig:(),gridData:(h:15,i:'1',w:24,x:0,y:0),id:'8f4d0c00-4c86-11e8-b3d7-01146121b73d',panelIndex:'1',type:visualization,version:'7.0.0-alpha1')),query:(language:lucene,query:''),timeRestore:!f,title:'New%20Dashboard',viewMode:edit)",
	}
	shortenURLResponse, err := client.API.KibanaShortenURL.Create(shortenURL)
	if err != nil {
		log.Fatalf("Error creating shorten URL: %s", err)
	}
	log.Println(fmt.Sprintf("http://localhost:5601/goto/%s", shortenURLResponse.ID))

	// Create or update Logstash pipeline
	logstashPipeline := &kbapi.LogstashPipeline{
		ID:          "sample",
		Description: "Sample logstash pipeline",
		Pipeline:    "input { stdin {} } output { stdout {} }",
		Settings: map[string]interface{}{
			"queue.type": "persisted",
		},
	}
	logstashPipeline, err = client.API.KibanaLogstashPipeline.CreateOrUpdate(logstashPipeline)
	if err != nil {
		log.Fatalf("Error creating logstash pipeline: %s", err)
	}
	log.Println(logstashPipeline)

	// Get the logstash pipeline
	logstashPipeline, err = client.API.KibanaLogstashPipeline.Get("sample")
	if err != nil {
		log.Fatalf("Error getting logstash pipeline: %s", err)
	}
	log.Println(logstashPipeline)

	// Get all logstash pipeline
	logstashPipelines, err := client.API.KibanaLogstashPipeline.List()
	if err != nil {
		log.Fatalf("Error getting all logstash pipeline: %s", err)
	}
	log.Println(logstashPipelines)

	// Delete logstash pipeline
	err = client.API.KibanaLogstashPipeline.Delete("sample")
	if err != nil {
		log.Fatalf("Error deleting logstash pipeline: %s", err)
	}
	log.Println("Logstash pipeline 'sample' successfully deleted")

	// Create user space
	space := &kbapi.KibanaSpace{
		ID:          "test",
		Name:        "test",
		Description: "My test",
	}
	space, err = client.API.KibanaSpaces.Create(space)
	if err != nil {
		log.Fatalf("Error creating user space: %s", err)
	}
	log.Println(space)

	// Update user space
	space.Name = "new name"
	space, err = client.API.KibanaSpaces.Update(space)
	if err != nil {
		log.Fatalf("Error updating user space: %s", err)
	}
	log.Println(space)

	// Get the user space
	space, err = client.API.KibanaSpaces.Get("test")
	if err != nil {
		log.Fatalf("Error getting user space: %s", err)
	}
	log.Println(space)

	// Get all user space
	spaces, err := client.API.KibanaSpaces.List()
	if err != nil {
		log.Fatalf("Error getting all user spaces: %s", err)
	}
	log.Println(spaces)

	// Copy config object from default space to test space
	parameter := &kbapi.KibanaSpaceCopySavedObjectParameter{
		Spaces:            []string{"test"},
		IncludeReferences: true,
		Overwrite:         true,
		Objects: []kbapi.KibanaSpaceObjectParameter{
			{
				Type: "config",
				ID:   "7.4.2",
			},
		},
	}
	err = client.API.KibanaSpaces.CopySavedObjects(parameter, "")
	if err != nil {
		log.Fatalf("Error copying object from another user space: %s", err)
	}
	log.Println("Copying config object from 'default' to 'test' user space successfully")

	// Delete user space
	err = client.API.KibanaSpaces.Delete("test")
	if err != nil {
		log.Fatalf("Error deleteing user space: %s", err)
	}
	log.Println("User space 'test' successfully deleted")

	// Import dashboard from file in default user space
	b, err := ioutil.ReadFile("../fixtures/kibana-dashboard.json")
	if err != nil {
		log.Fatalf("Error reading file: %s", err)
	}
	data := make(map[string]interface{})
	err = json.Unmarshal(b, &data)
	err = client.API.KibanaDashboard.Import(data, nil, true, "default")
	if err != nil {
		log.Fatalf("Error importing dashboard: %s", err)
	}
	log.Println("Importing dashboard successfully")

	// Export dashboard from default user space
	data, err = client.API.KibanaDashboard.Export([]string{"edf84fe0-e1a0-11e7-b6d5-4dc382ef7f5b"}, "default")
	if err != nil {
		log.Fatalf("Error exporting dashboard: %s", err)
	}
	log.Println(data)

	// Create or update role
	role := &kbapi.KibanaRole{
		Name: "test",
		Elasticsearch: &kbapi.KibanaRoleElasticsearch{
			Indices: []kbapi.KibanaRoleElasticsearchIndice{
				{
					Names: []string{
						"*",
					},
					Privileges: []string{
						"read",
					},
				},
			},
		},
		Kibana: []kbapi.KibanaRoleKibana{
			{
				Base: []string{
					"read",
				},
			},
		},
	}
	role, err = client.API.KibanaRoleManagement.CreateOrUpdate(role)
	if err != nil {
		log.Fatalf("Error creating role: %s", role)
	}
	log.Println(role)

	// Get the role
	role, err = client.API.KibanaRoleManagement.Get("test")
	if err != nil {
		log.Fatalf("Error reading role: %s", role)
	}
	log.Println(role)

	// List all roles
	roles, err := client.API.KibanaRoleManagement.List()
	if err != nil {
		log.Fatalf("Error reading all roles: %s", err)
	}
	log.Println(roles)

	// Delete role
	err = client.API.KibanaRoleManagement.Delete("test")
	if err != nil {
		log.Fatalf("Error deleting role: %s", err)
	}
	log.Println("Role successfully deleted")

	// Create new index pattern in default user space
	dataJSON := `{"attributes": {"title": "test-pattern-*"}}`
	data = make(map[string]interface{})
	err = json.Unmarshal([]byte(dataJSON), &data)
	if err != nil {
		log.Fatalf("Error converting json to struct: %s", err)
	}
	resp, err := client.API.KibanaSavedObject.Create(data, "index-pattern", "test", true, "default")
	if err != nil {
		log.Fatalf("Error creating object: %s", err)
	}
	log.Println(resp)

	// Get index pattern save object from default user space
	resp, err = client.API.KibanaSavedObject.Get("index-pattern", "test", "default")
	if err != nil {
		log.Fatalf("Error getting index pattern save object: %s", err)
	}
	log.Println(resp)

	// Search index pattern from default user space
	parameters := &kbapi.OptionalFindParameters{
		Search:       "test",
		SearchFields: []string{"id"},
		Fields:       []string{"id"},
	}
	resp, err = client.API.KibanaSavedObject.Find("index-pattern", "default", parameters)
	if err != nil {
		log.Fatalf("Error searching index pattern: %s", err)
	}
	log.Println(resp)

	// Update index pattern in default user space
	dataJSON = `{"attributes": {"title": "test-pattern2-*"}}`
	err = json.Unmarshal([]byte(dataJSON), &data)
	if err != nil {
		log.Fatalf("Error converting json to struct")
	}
	resp, err = client.API.KibanaSavedObject.Update(data, "index-pattern", "test", "default")
	if err != nil {
		log.Fatalf("Error updating index pattern: %s", err)
	}

	// Export index pattern from default user space
	request := []map[string]string{
		{
			"type": "index-pattern",
			"id":   "test",
		},
	}
	response, error := client.API.KibanaSavedObject.Export(nil, request, true, "default")
	if error != nil {
		log.Fatalf("Error exporting index pattern: %s", error)
	}
	log.Println(response)

	// import index pattern in default user space
	b, err = json.Marshal(response)
	if err != nil {
		log.Fatalf("Error converting struct to json")
	}
	resp2, err := client.API.KibanaSavedObject.Import(b, true, "default")
	if err != nil {
		log.Fatalf("Error importing index pattern: %s", err)
	}
	log.Println(resp2)

	// Delete index pattern in default user space
	err = client.API.KibanaSavedObject.Delete("index-pattern", "test", "default")
	if err != nil {
		log.Fatalf("Error deleting index pattern: %s", err)
	}
	log.Println("Index pattern successfully deleted")

}
