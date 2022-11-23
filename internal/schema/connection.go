package schema

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func GetConnectionResource(keyName string) *schema.Resource {
	username := makePathRef(keyName, "username")
	password := makePathRef(keyName, "password")
	caFile := makePathRef(keyName, "ca_file")
	caData := makePathRef(keyName, "ca_data")
	certFile := makePathRef(keyName, "cert_file")
	certData := makePathRef(keyName, "cert_data")
	keyFile := makePathRef(keyName, "key_file")
	keyData := makePathRef(keyName, "key_data")

	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"username": {
				Description:  "A username to use for API authentication to Elasticsearch.",
				Type:         schema.TypeString,
				Optional:     true,
				RequiredWith: []string{password},
			},
			"password": {
				Description:  "A password to use for API authentication to Elasticsearch.",
				Type:         schema.TypeString,
				Optional:     true,
				Sensitive:    true,
				RequiredWith: []string{username},
			},
			"api_key": {
				Description:   "API Key to use for authentication to Elasticsearch",
				Type:          schema.TypeString,
				Optional:      true,
				Sensitive:     true,
				ConflictsWith: []string{username, password},
			},
			"endpoints": {
				Description: "A list of endpoints the Terraform provider will point to. They must include the http(s) schema and port number.",
				Type:        schema.TypeList,
				Optional:    true,
				Sensitive:   true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"insecure": {
				Description: "Disable TLS certificate validation",
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ELASTICSEARCH_INSECURE", false),
			},
			"ca_file": {
				Description:   "Path to a custom Certificate Authority certificate",
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{caData},
			},
			"ca_data": {
				Description:   "PEM-encoded custom Certificate Authority certificate",
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{caFile},
			},
			"cert_file": {
				Description:   "Path to a file containing the PEM encoded certificate for client auth",
				Type:          schema.TypeString,
				Optional:      true,
				RequiredWith:  []string{keyFile},
				ConflictsWith: []string{certData, keyData},
			},
			"key_file": {
				Description:   "Path to a file containing the PEM encoded private key for client auth",
				Type:          schema.TypeString,
				Optional:      true,
				RequiredWith:  []string{certFile},
				ConflictsWith: []string{certData, keyData},
			},
			"cert_data": {
				Description:   "PEM encoded certificate for client auth",
				Type:          schema.TypeString,
				Optional:      true,
				RequiredWith:  []string{keyData},
				ConflictsWith: []string{certFile, keyFile},
			},
			"key_data": {
				Description:   "PEM encoded private key for client auth",
				Type:          schema.TypeString,
				Optional:      true,
				Sensitive:     true,
				RequiredWith:  []string{certData},
				ConflictsWith: []string{certFile, keyFile},
			},
		},
	}
}

func makePathRef(keyName string, keyValue string) string {
	return fmt.Sprintf("%s.0.%s", keyName, keyValue)
}
