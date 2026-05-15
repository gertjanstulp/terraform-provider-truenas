# Retrieve an existing SSH keypair
data "truenas_ssh_keypair" "keypair" {
  name = "keypair"
}

# Create a new SSH connection using the keypair
resource "truenas_ssh_connection" "connection" {
  name = "connection"
  host = "example.com"
  port = 22
  username = "username"
  private_key_id = data.truenas_ssh_keypair.keypair.id
  remote_host_key = "<some key>"
  connect_timeout = 30
}