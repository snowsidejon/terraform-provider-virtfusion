resource "virtfusion_build" "node1" {
  server_id = virtfusion_server.node1.id
  name      = "node1-demo"
  hostname  = "node1.example.com"
  osid      = 1
  vnc       = true
  ipv6      = true
  ssh_keys  = [virtfusion_ssh.dummy_key.id]
  email     = true
}
