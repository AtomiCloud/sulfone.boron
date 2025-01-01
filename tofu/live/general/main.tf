data "cloudflare_zone" "this" {
  name = "cyanprint.dev"
}

resource "cloudflare_record" "this" {
  zone_id = data.cloudflare_zone.this.id
  name    = "coord"
  content   = var.ip_address
  type    = "A"
  ttl     = 3600
  proxied = false
}