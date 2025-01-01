{ pkgs, packages }:
with packages;
{
  system = [
    coreutils
    sd
    bash
    findutils
  ];

  dev = [
    pls
    git
  ];

  infra = [
    docker
  ];

  main = [
    gcc
    go
    infisical
    opentofu
    ansible
  ];

  lint = [
    # core
    treefmt
    hadolint
    gitlint
    shellcheck
    sg
    golangci-lint
  ];

  ci = [

  ];

  releaser = [
    sg
  ];

}
