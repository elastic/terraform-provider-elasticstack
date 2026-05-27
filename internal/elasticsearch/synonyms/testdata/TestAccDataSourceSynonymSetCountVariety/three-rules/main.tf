variable "synonym_set_id" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_synonym_set" "src" {
  synonym_set_id = var.synonym_set_id

  synonyms_set = [
    {
      id       = "rule-a"
      synonyms = "cat, feline"
    },
    {
      id       = "rule-b"
      synonyms = "dog, hound, canine"
    },
    {
      id       = "rule-c"
      synonyms = "happy, joyful, cheerful"
    },
  ]
}

data "elasticstack_elasticsearch_synonym_set" "test" {
  synonym_set_id = elasticstack_elasticsearch_synonym_set.src.synonym_set_id
}
