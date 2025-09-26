terraform {
  required_providers {
    virtfusion = {
      source  = "snowsidejon/virtfusion"
      version = "1.0.2"
    }
  }
}

# Default endpoint = cloud.breezehost.io
# Default values for resource_package, os_template, etc. come from env vars.
provider "virtfusion" {
  api_token = var.api_token
}

variable "api_token" {
  type      = string
  sensitive = true
}
