# VirtFusion Terraform Provider

![Terraform](https://img.shields.io/badge/Terraform-%235835CC.svg?style=for-the-badge&logo=terraform&logoColor=white)
![Go](https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white)

A Terraform provider for [VirtFusion](https://virtfusion.com), maintained by [BreezeHost](https://breezehost.io).  
Provision and manage VPS instances directly from Terraform.

---

## ✨ Features
- 🔑 Authenticate with API tokens  
- ☁️ Default endpoint: `https://cloud.breezehost.io`  
- 📦 Deploy VMs with resource packages (CPU/RAM/Disk plans)  
- 💾 Select OS templates automatically (defaults to Ubuntu 22.04)  
- 🌍 Assign public and private IPs  
- 🏢 Deploy automatically into hypervisor groups (locations)  
- 🔒 Manage SSH keys and builds  

---

## 🚀 Quick Start

### Install Provider
```hcl
terraform {
  required_providers {
    virtfusion = {
      source  = "snowsidejon/virtfusion"
      version = "1.0.2"
    }
  }
}

### Configure via Environment Variables

Set your credentials and defaults:

export VIRTFUSION_API_TOKEN="your_api_token"
export VIRTFUSION_OS_TEMPLATE="Ubuntu Server 22.04"
export VIRTFUSION_RESOURCE_PACKAGE=11
export VIRTFUSION_PUBLIC_IPS=1
export VIRTFUSION_PRIVATE_IPS=0
export VIRTFUSION_HYPERVISOR_GROUP=14


### Or, on PowerShell:

$env:VIRTFUSION_API_TOKEN="your_api_token"
$env:VIRTFUSION_RESOURCE_PACKAGE="11"

## 🛠 Example Usage
Create a VM
provider "virtfusion" {}

resource "virtfusion_server" "vm" {
  name              = "terraform-vm"
  os_template       = "Debian 12"
  resource_package  = 15
  public_ips        = 2
  private_ips       = 1
  hypervisor_group  = 14
}

## Add SSH Key
resource "virtfusion_ssh" "key1" {
  name       = "My Test Key"
  public_key = "ssh-ed25519 AAAAC3Nz..."
}

## Build VM with OS + SSH
resource "virtfusion_build" "vm_build" {
  server_id   = virtfusion_server.vm.id
  name        = "node1"
  hostname    = "node1.example.com"
  osid        = 34
  ssh_keys    = [virtfusion_ssh.key1.id]
  vnc         = true
  ipv6        = true
  email       = true
}

## ⚙️ Configuration Options
Argument	Env Var	Default	Description
endpoint	VIRTFUSION_ENDPOINT	cloud.breezehost.io	API endpoint
api_token	VIRTFUSION_API_TOKEN	Required	API token
os_template	VIRTFUSION_OS_TEMPLATE	Ubuntu Server 22.04	OS template
resource_package	VIRTFUSION_RESOURCE_PACKAGE	1	Resource package ID
public_ips	VIRTFUSION_PUBLIC_IPS	1	Number of public IPs
private_ips	VIRTFUSION_PRIVATE_IPS	0	Number of private IPs
hypervisor_group	VIRTFUSION_HYPERVISOR_GROUP	1	Hypervisor group (location) ID
📝 Contributing

PRs welcome! Please open issues or submit patches if you’d like to extend functionality.

📖 License

MPL-2.0 License © 2025 BreezeHost


---

✅ With this, your provider:
- Defaults to BreezeHost cloud.  
- Supports env vars for all the “knobs” (OS, package, IPs, hypervisor group).  
- Has a clean README for new users.  

---

Do you want me to also draft an **`examples/` folder structure** (`examples/basic`, `examples/multi-ip`, `