#!/usr/bin/env bash

set -eou pipefail

find . -type f -name "0_statics.tf" ! -path "./0_statics.tf" -exec cp ./0_statics.tf {} \;

folders=$(find . -type f -name "0_statics.tf" ! -path "./0_statics.tf" -exec dirname {} \;)

# Copy the file "./*.auto.tfvars" to each found folder
for folder in $folders; do
  cp ./*.auto.tfvars "$folder"
done
