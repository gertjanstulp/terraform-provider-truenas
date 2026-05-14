# Retrieve information about an existing SSH keypair
data "truenas_sshkeypair" "keypair" {
  name = "keypair"
}

output "sshkeypair_id" {
  value = data.truenas_sshkeypair.keypair.id
}
