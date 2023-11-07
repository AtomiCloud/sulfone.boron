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
    helm
    kubectl
    docker
  ];

  main = [
    go
    infisical
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
    nodejs
    sg
    npm
  ];

}
