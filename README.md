# VirtFusion Terraform Provider

![Terraform](https://img.shields.io/badge/Terraform-%235835CC.svg?style=for-the-badge&logo=terraform&logoColor=white)
![Go](https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white)
![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/snowsidejon/terraform-provider-virtfusion)

## Overview

The VirtFusion Terraform provider allows you to manage BreezeHost VirtFusion resources with Terraform.  
It supports creating servers, provisioning builds, and managing SSH keys.

- 🚀 Automatic defaults (endpoint, resource packages, OS templates, IP counts)  
- 🔑 Environment variable support for easy automation  
- 🌍 Community provider on the [Terraform Registry](https://registry.terraform.io/providers/snowsidejon/virtfusion/latest)  

---

## Installation

```hcl
terraform {
  required_providers {
    virtfusion = {
      source  = "snowsidejon/virtfusion"
      version = "1.0.3"
    }
  }
}
```

---

## Provider Configuration

The provider can be configured using HCL attributes or environment variables.  

### Attributes
```hcl
provider "virtfusion" {
  endpoint         = "cloud.breezehost.io"
  api_token        = var.api_token
  os_template      = "Ubuntu Server 22.04"
  resource_package = 11
  public_ips       = 1
  private_ips      = 0
  hypervisor_group = 14
}
```

### Environment variables
| Attribute         | Env Var                       | Default                  |
|-------------------|-------------------------------|--------------------------|
| `endpoint`        | `VIRTFUSION_ENDPOINT`         | `cloud.breezehost.io`    |
| `api_token`       | `VIRTFUSION_API_TOKEN`        | _none (required)_        |
| `os_template`     | `VIRTFUSION_OS_TEMPLATE`      | `Ubuntu Server 22.04`    |
| `resource_package`| `VIRTFUSION_RESOURCE_PACKAGE` | n/a                      |
| `public_ips`      | `VIRTFUSION_PUBLIC_IPS`       | `1`                      |
| `private_ips`     | `VIRTFUSION_PRIVATE_IPS`      | `0`                      |
| `hypervisor_group`| `VIRTFUSION_HYPERVISOR_GROUP` | n/a                      |

---

## Example: Basic

```hcl
provider "virtfusion" {}

resource "virtfusion_server" "demo" {
  user_id = 1
}

resource "virtfusion_build" "demo" {
  server_id = virtfusion_server.demo.id
  name      = "tf-basic"
  hostname  = "tf-basic.example.com"
}
```

Just set your API token:

```bash
export VIRTFUSION_API_TOKEN="your_api_token"
terraform init
terraform apply
```

---

## Example: Advanced

```hcl
provider "virtfusion" {
  api_token        = var.api_token
  os_template      = "Debian 12"
  resource_package = 15
  public_ips       = 2
  private_ips      = 1
  hypervisor_group = 14
}

variable "api_token" {
  type      = string
  sensitive = true
}

resource "virtfusion_ssh" "my_key" {
  user_id    = 1
  name       = "terraform-key"
  public_key = "ssh-ed25519 AAAAC3NzExampleKeyGeneratedLocally"
}

resource "virtfusion_server" "vm" {
  user_id = 1
}

resource "virtfusion_build" "vm" {
  server_id = virtfusion_server.vm.id
  name      = "adv-vm"
  hostname  = "adv.example.com"
  ssh_keys  = [virtfusion_ssh.my_key.id]
  vnc       = true
  ipv6      = true
  email     = true
}
```

---

## Resources

- `virtfusion_server` → Create and manage virtual machines  
- `virtfusion_build` → Provision and configure servers  
- `virtfusion_ssh` → Manage SSH keys  

---

## Contributing

Issues and PRs are welcome. Please fork, branch, and submit a PR with changes.  
For security-sensitive issues, contact BreezeHost directly.

---

## License

MPL-2.0 — see [LICENSE](LICENSE).
