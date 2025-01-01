
provider "infisical" {
  host          = "https://secrets.atomi.cloud"
  client_id     = var.infisical_client_id
  client_secret = var.infisical_client_secret
}

data "infisical_projects" "sulfone_raichu" {
  slug  = "${local.platform.slug}-${local.service.slug}"
}

data "infisical_secrets" "sulfone_raichu" {
  env_slug     = local.landscape.slug
  workspace_id = data.infisical_projects.sulfone_raichu.id
  folder_path  = "/"
}

provider "cloudflare" {
  api_token = data.infisical_secrets.sulfone_raichu.secrets["CLOUDFLARE_API_TOKEN"].value
}
