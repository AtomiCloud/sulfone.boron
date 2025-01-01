terraform {

  backend "pg" {
    conn_str    = var.tofu_backend
    schema_name = "general-${local.service.slug}"
  }

  required_providers {
    cloudflare = {
      source  = "cloudflare/cloudflare"
      version = "~> 4"
    }
    infisical = {
      source = "Infisical/infisical"
      version = "~> 0"
    }
  }
}
