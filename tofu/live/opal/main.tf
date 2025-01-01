module "compute" {
  source = "../../modules/compute/digitalocean"

  landscape = local.landscape.slug
  platform  = local.platform.slug
  service   = local.service.slug
  module    = "compute"

  instance_type = local.size
  user          = "kirin"
  ssh_key = [
    "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQDy0SiuIqB3Zwj1dFb5Khysa9gGSBeA6IYHREPBNPWZUWDV8WqKCeuMNebDC2USQcAlF/ewPhsz0bGq5k4PqmnmbSHWiM1KI7BJkYolLcv/upR3a5R2UGoNdVXA0AlXiFORNxM444NhRIOx18wfrHbtRXBfdJUlWx9gHqkeMEQ1OabVDHd+C8nDFvwe6kAplJOyoaqZ0l8LAgSZlgE7GyZjgkHLDSMI1PXBEJu7JB61mNt2FAYGrdnolA0Ryv7WWWLThOoFq1v7aKX1M1bA71AmV3e//HOBZ945eYEMj5z+cwdRUxeszUdGV0HHe0/qDu98Pd7meJqitjWbuDJ3wX5HbMZ8ctO2cH5Z82Qc9DK6xylWHduBefyNGP1q9SE6mRiLU5QrFHDeh0irZqh6NIhQyXqg9T/25f94TnM4WD5i1EPpoj4OfMpWd5qmo7ZVQVVNsDkaIs0K+ncwVddeLAXuiUOzxeyA6tKvcnTXYX8sT+/qrmVRPRpsN9kDjuYkskfXNuBqyA3QID0jvS0N6vHQx0HKMS8R4qLYZpu5Qh1A2/GOPfSNwsSOBRLO57lX/0NP/lLY/+u3pP4n8RZVpxeVu1oOZoq78ujEoLWygyA6RWFIbuYy6mDe1gFrVNrkOfYV0dBmF1xEcmOOryIl+aPRsv8iRat3unw5NZywiDG3Lw== ernest@MacBook-Pro",
  ]
  region = "sgp1"
}
