locals {

  landscape = local.landscapes.raichu
  platform = local.platforms.sulfone
  service = local.platforms.sulfone.services.boron

  cluster = local.clusters.digital_ocean.green

  size = "s-2vcpu-4gb"
}