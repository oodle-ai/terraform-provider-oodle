terraform {
  required_providers {
    oodle = {
      source = "registry.terraform.io/oodle-ai/oodle"
    }
  }
}

provider "oodle" {
  deployment_url = "https://us1.oodle.ai/"
  instance       = "my-instance"
  api_key        = "my-api-key"
}

