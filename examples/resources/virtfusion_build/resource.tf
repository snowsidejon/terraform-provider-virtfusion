resource "virtfusion_server" "node1" {
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
}


resource "virtfusion_build" "node1" {
  server_id = virtfusion_server.node1.id
  name      = "node1-demo"
  hostname  = "node1.example.com"
  osid      = 1
  vnc       = true
  ipv6      = true
  ssh_keys  = [1, 2, 3]
  email     = true
}
