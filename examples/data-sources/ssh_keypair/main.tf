# Retrieve information about an existing SSH keypair
data "truenas_ssh_keypair" "keypair" {
  name = "keypair"
}

output "ssh_keypair_id" {
  value = data.truenas_ssh_keypair.keypair.id
}
