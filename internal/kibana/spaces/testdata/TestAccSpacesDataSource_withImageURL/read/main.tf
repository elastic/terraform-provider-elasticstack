variable "space_id" {
  description = "The space ID for the custom test space"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_space" "test" {
  space_id  = var.space_id
  name      = "Test Image URL Space"
  image_url = "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg=="
}

data "elasticstack_kibana_spaces" "all_spaces" {
  depends_on = [elasticstack_kibana_space.test]
}
