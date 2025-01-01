locals {

  lpsm      = "${var.landscape}-${var.platform}-${var.service}-${var.module}"
  lpsm-fqdn = "${var.landscape}.${var.platform}.${var.service}.${var.module}"

  tags = {
    Name      = local.lpsm
    FQDN      = local.lpsm-fqdn
    ManagedBy = "tofu"
  }

  os_image = "ubuntu-24-04-x64"
}