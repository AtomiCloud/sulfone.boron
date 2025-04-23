{ pkgs, atomi, pkgs-2411, pkgs-2405 }:
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
      with pkgs;
      { }
    );
    nix-2405 = (
      with pkgs-2405;
      {
        inherit
          ansible;
      }
    );
    nix-2411 = (
      with pkgs-2411;
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
nix-2411 //
nix-2405 //
nix-unstable
