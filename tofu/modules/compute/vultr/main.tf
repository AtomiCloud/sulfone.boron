

data "vultr_os" "ubuntu" {
  filter {
    name   = "name"
    values = [local.os_image]
  }
}

resource "vultr_instance" "this" {
  os_id  = data.vultr_os.ubuntu.id
  label  = local.lpsm
  region = var.region
  plan   = var.instance_type

  user_data = templatefile("${path.module}/cloud-init.yaml.tpl", {
    ssh_keys = var.ssh_key
    user     = var.user
  })
}