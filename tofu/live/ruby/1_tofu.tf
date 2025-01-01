terraform {

  backend "pg" {
    conn_str    = var.tofu_backend
    schema_name = "${local.landscape.slug}-${local.cluster}/${local.platform.slug}-${local.service.slug}"
  }


  required_providers {
    digitalocean = {
      source  = "digitalocean/digitalocean"
      version = ">= 2"
    }
    infisical = {
      source = "Infisical/infisical"
      version = "~> 0"
    }
  }
}
