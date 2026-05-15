# Generate a new public/private keypair
resource "tls_private_key" "keypair" {
  algorithm = "RSA"
  rsa_bits = 2048
}

# Create a new truenas keypair
resource "truenas_ssh_keypair" "keypair" {
  name = "keypair"
  public_key = tls_private_key.keypair.public_key_openssh
  private_key = tls_private_key.keypair.private_key_openssh
}