# Data source to fetch the latest Ubuntu AMI
data "aws_ami" "latest_ubuntu" {
  most_recent = true

  owners = ["099720109477"]  # Official Ubuntu AMI owner

  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd/ubuntu*-arm64-*"]
  }
  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }
}