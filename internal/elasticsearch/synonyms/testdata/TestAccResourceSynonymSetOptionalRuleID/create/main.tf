variable "synonym_set_id" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_synonym_set" "test" {
  synonym_set_id = var.synonym_set_id

  synonyms_set = [
    {
      synonyms = "quick, fast, speedy"
    },
    {
      id       = "explicit-rule"
      synonyms = "slow, sluggish"
    },
  ]
}
