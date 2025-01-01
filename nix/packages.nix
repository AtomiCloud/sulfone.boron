{ pkgs, atomi, pkgs-2411 }:
let
  all = {
    atomipkgs = (
      with atomi;
      {
        inherit
          sg
          pls;
      }
    );
    nix-2411 = (
      with pkgs-2411;
      {
        inherit
          ansible
          infisical
          hadolint
          k3d
          coreutils
          findutils
          sd
          bash
          git

          gcc
          opentofu

          # lint
          treefmt
          gitlint
          shellcheck

          go
          golangci-lint

          #infra
          kubectl
          docker;
      }
    );

  };
in
with all;
atomipkgs //
nix-2411
