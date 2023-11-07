{ pkgs, atomi, atomi_classic, pkgs-2305, pkgs-sep-04-23 }:
let
  all = {
    atomipkgs_classic = (
      with atomi_classic;
      {
        inherit
          sg;
      }
    );
    atomipkgs = (
      with atomi;
      {
        inherit
          infisical
          pls;
      }
    );
    nix-2305 = (
      with pkgs-2305;
      {
        inherit
          hadolint
          k3d;
      }
    );
    sep-04-23 = (
      with pkgs-sep-04-23;
      {
        inherit
          coreutils
          findutils
          sd
          bash
          git
          go

          # lint
          treefmt
          gitlint
          shellcheck
          golangci-lint

          #infra
          kubectl
          docker;
        helm = kubernetes-helm;
        npm = nodePackages.npm;
        nodejs = nodejs_20;
      }
    );
  };
in
with all;
atomipkgs //
atomipkgs_classic //
nix-2305 //
sep-04-23
