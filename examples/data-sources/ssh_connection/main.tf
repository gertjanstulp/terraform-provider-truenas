# Retrieve information about an existing SSH connection
data "truenas_ssh_connection" "connection" {
  name = "connection"
}

output "ssh_connection_id" {
  value = data.truenas_ssh_connection.connection.id
}
