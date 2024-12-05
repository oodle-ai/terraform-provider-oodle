terraform {
  required_providers {
    oodle = {
      source = "registry.terraform.io/hashicorp/oodle"
    }
  }
}

provider "oodle" {}

resource "oodle_monitor" "test1" {
  # (resource arguments)
}
