---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "oodle Provider"
subcategory: ""
description: |-
  
---

# oodle Provider



## Example Usage

```terraform
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
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `api_key` (String, Sensitive)
- `deployment_url` (String)
- `instance` (String)
