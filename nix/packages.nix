{ pkgs, atomi, pkgs-2505, pkgs-2405, pkgs-unstable }:
let
  all = {
    atomipkgs = (
      with atomi;
      {
        inherit
          atomiutils
          infrautils
          infralint
          sg
          pls;
      }
    );
    nix-unstable = (
      with pkgs-unstable;
      { }
    );
    nix-2405 = (
      with pkgs-2405;
      {
        inherit
          ansible;
      }
    );
    nix-2505 = (
      with pkgs-2505;
      {
        inherit
          infisical
          git

          gcc

          # lint
          treefmt
          gitlint
          shellcheck

          go
          golangci-lint;
      }
    );

  };
in
with all;
atomipkgs //
nix-2505 //
nix-2405 //
nix-unstable
