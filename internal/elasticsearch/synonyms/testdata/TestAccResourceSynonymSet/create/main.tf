variable "synonym_set_id" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_synonym_set" "test" {
  synonym_set_id = var.synonym_set_id

  synonyms_set {
    id       = "rule-1"
    synonyms = "i-pod, i pod => ipod"
  }

  synonyms_set {
    id       = "rule-2"
    synonyms = "universe, cosmos"
  }
}
