# tfsec:ignore:aws-ec2-enable-at-rest-encryption tfsec:ignore:aws-ec2-enforce-http-token-imds
resource "aws_instance" "this" {
  ami           = data.aws_ami.latest_ubuntu.id
  instance_type = var.instance_type
  subnet_id     = aws_subnet.this.id

  user_data = templatefile("${path.module}/cloud-init.yaml.tpl", {
    ssh_keys = var.ssh_key
    user = var.user
  })

  tags = local.tags
  vpc_security_group_ids = [aws_security_group.this.id]
  associate_public_ip_address = true
}
