#cloud-config

users:
  - name: ${user}
    sudo: ALL=(ALL) NOPASSWD:ALL
    groups: users, admin, docker
    lock_passwd: true
    ssh_authorized_keys:
      %{ for key in ssh_keys ~}
      - ${key}
      %{ endfor ~}

groups:
  - docker

system_info:
  default_user:
    groups: [docker]
