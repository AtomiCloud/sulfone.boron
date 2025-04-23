{ pkgs, packages }:
with packages;
{
  system = [
    atomiutils
  ];

  dev = [
    pls
    git
  ];

  infra = [
  ];

  main = [
    gcc
    go
    infisical
    infrautils
    ansible
  ];

  lint = [
    # core
    treefmt
    gitlint
    shellcheck
    sg
    golangci-lint
    infralint
  ];

  ci = [

  ];

  releaser = [
    sg
  ];

}
