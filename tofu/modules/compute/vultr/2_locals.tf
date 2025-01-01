locals {

  lpsm      = "${var.landscape}-${var.platform}-${var.service}-${var.module}"
  lpsm-fqdn = "${var.landscape}.${var.platform}.${var.service}.${var.module}"

  tags = {
    Name      = local.lpsm
    FQDN      = local.lpsm-fqdn
    ManagedBy = "tofu"
  }

  os_image = "Ubuntu 24.04 LTS x64"
}