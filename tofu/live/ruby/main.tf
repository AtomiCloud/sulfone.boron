module "compute" {
  source = "../../modules/compute/digitalocean"

  landscape = local.landscape.slug
  platform  = local.platform.slug
  service   = local.service.slug
  module    = "compute"

  instance_type = local.size
  user          = "kirin"
  ssh_key = [
    "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAINMIrixL20JiRNhFF2ziG2Ar4aJwImKq5Qq2je4FSGFL ernest@Ernests-MBP",
    "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIAxyfAwdu+4Lk9wbJp7qb0GtDoiSe7utUk20jsrao12e ernest@Ernests-MBP",
  ]
  region = "sgp1"
}
