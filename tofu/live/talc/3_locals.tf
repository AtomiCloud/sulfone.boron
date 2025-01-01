locals {

  landscape = local.landscapes.raichu
  platform = local.platforms.sulfone
  service = local.platforms.sulfone.services.boron

  cluster = local.clusters.vultr.green

  # curl "https://api.vultr.com/v2/plans" | jq '.plans[].id'
  size = "vc2-1c-1gb"
}