variable "tofu_backend" {
  type = string
  # sensitive = true
}

# secrets
variable "infisical_client_id" {
  type      = string
  sensitive = true
}
variable "infisical_client_secret" {
  type      = string
  sensitive = true
}