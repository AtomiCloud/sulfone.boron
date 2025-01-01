# tfsec:ignore:digitalocean-compute-use-ssh-keys
resource "digitalocean_droplet" "this" {
  image    = local.os_image
  name     = local.lpsm
  region   = var.region
  size     = var.instance_type
  ssh_keys = []
  user_data = templatefile("${path.module}/cloud-init.yaml.tpl", {
    ssh_keys = var.ssh_key
    user     = var.user
  })
}