resource "virtfusion_example" "server" {
  package_id             = 1
  user_id                = 1
  hypervisor_id          = 1
  ipv4                   = 1
  storage                = 30
  memory                 = 1024
  cores                  = 1
  traffic                = 1000
  inbound_network_speed  = 100
  outbound_network_speed = 100
  storage_profile        = 1
  network_profile        = 1
  name                   = "test"
}
