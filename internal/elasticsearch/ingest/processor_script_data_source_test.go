package ingest_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceIngestProcessorScript(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceIngestProcessorScript,
				Check: resource.ComposeTestCheckFunc(
					CheckResourceJson("data.elasticstack_elasticsearch_ingest_processor_script.test", "json", expectedJsonScript),
				),
			},
		},
	})
}

const expectedJsonScript = `{
	"script": {
		"description": "Extract 'tags' from 'env' field",
		"ignore_failure": false,
		"lang": "painless",
		"params": {
			"delimiter": "-",
			"position": 1
		},
		"source": "String[] envSplit = ctx['env'].splitOnToken(params['delimiter']);\nArrayList tags = new ArrayList();\ntags.add(envSplit[params['position']].trim());\nctx['tags'] = tags;\n"
	}
}`

const testAccDataSourceIngestProcessorScript = `
provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_script" "test" {
  description = "Extract 'tags' from 'env' field"
  lang        = "painless"

  source = <<EOF
String[] envSplit = ctx['env'].splitOnToken(params['delimiter']);
ArrayList tags = new ArrayList();
tags.add(envSplit[params['position']].trim());
ctx['tags'] = tags;
EOF

  params = jsonencode({
    delimiter = "-"
    position  = 1
  })

}
`
