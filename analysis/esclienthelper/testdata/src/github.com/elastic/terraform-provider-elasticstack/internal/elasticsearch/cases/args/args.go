package args

import (
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func firstParamSink(d *schema.ResourceData, meta any) error {
	client, _ := clients.NewAPIClientFromSDKResource(d, meta)
	return elasticsearch.DoFirst(client, "ok")
}

func secondParamSink(d *schema.ResourceData, meta any) error {
	client, _ := clients.NewAPIClientFromSDKResource(d, meta)
	return elasticsearch.DoSecond("ok", client)
}

func secondParamViolation() error {
	client := &clients.APIClient{}
	return elasticsearch.DoSecond("bad", client) // want "Elasticsearch client usage must use a helper-derived \\*clients.APIClient from clients.NewAPIClientFromSDKResource\\(\\.\\.\\.\\) or clients.MaybeNewAPIClientFromFrameworkResource\\(\\.\\.\\.\\)"
}

func firstParamViolation() error {
	client := &clients.APIClient{}
	return elasticsearch.DoFirst(client, "bad") // want "Elasticsearch client usage must use a helper-derived \\*clients.APIClient from clients.NewAPIClientFromSDKResource\\(\\.\\.\\.\\) or clients.MaybeNewAPIClientFromFrameworkResource\\(\\.\\.\\.\\)"
}
