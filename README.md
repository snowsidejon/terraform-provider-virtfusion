# Virtfusion Terraform Provider

<p style="color: red">NOTE: This is a work in progress and is not yet ready for production use.</p>


## Overview

This is a Terraform provider for the Virtfusion API. It allows you to manage your Virtfusion resources using Terraform.

# What can I do with this provider?

Currently, you're able to manage the following resources:
* Create and delete virtual machines
* Create and delete SSH keys

# How do I use this provider?

Below is an example of how to use this provider to create a virtual machine and an SSH key.

```hcl
terraform {
  required_providers {
    virtfusion = {
      source = "ezscale/virtfusion"
        version = "0.0.3"
    }
  }
}

provider "virtfusion" {
  endpoint = "virtfusion.example.com"
  api_token = ""
}

variable "common" {
    type = map(string)
    default = {
        hypervisor_id = 1
        package_id = 12
        user_id = 1
    }
}

# Create a SSH key
resource "virtfusion_ssh" "key1" {
  name = "My Test Key"
  public_key = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIKWyBR+dk5M5MMfmH6Ss5QDSgcAvbCYu0DkqgPKH8O5T testkey@example.com"
  user_id = var.common["user_id"]
}

# Create a server
resource "virtfusion_server" "node1" {
  hypervisor_id = var.common["hypervisor_id"]
  package_id = var.common["package_id"]
  user_id = var.common["user_id"]
}

# Initialize the server with the OS we want, the SSH key we want, and the hostname we want.
resource "virtfusion_build" "node1" {
  server_id = virtfusion_server.node1.id
  name = "node1-demo"
  hostname = "node1.example.com"
  osid = 34
  vnc = true
  ipv6 = true
  ssh_keys = [virtfusion_ssh.key1.id]
  email = true
}
```

# How can I contribute?

If you'd like to contribute, please feel free to open a pull request. If you're unsure of what to work on, please check the issues tab for any open issues.