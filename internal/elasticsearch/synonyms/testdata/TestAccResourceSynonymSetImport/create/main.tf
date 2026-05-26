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
      id       = "rule-a"
      synonyms = "dog, hound, canine"
    },
    {
      id       = "rule-b"
      synonyms = "cat, feline"
    },
  ]
}
