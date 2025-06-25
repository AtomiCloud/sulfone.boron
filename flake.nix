{
  inputs = {
    # util
    flake-utils.url = "github:numtide/flake-utils";
    treefmt-nix.url = "github:numtide/treefmt-nix";
    pre-commit-hooks.url = "github:cachix/pre-commit-hooks.nix";

    # registry
    nixpkgs-unstable.url = "nixpkgs/nixos-unstable";
    nixpkgs-2505.url = "nixpkgs/nixos-25.05";
    nixpkgs-2405.url = "nixpkgs/nixos-24.05";
    atomipkgs.url = "github:AtomiCloud/nix-registry/v2";
  };
  outputs =
    { self

      # utils
    , flake-utils
    , treefmt-nix
    , pre-commit-hooks

      # registries
    , atomipkgs
    , nixpkgs-unstable
    , nixpkgs-2505
    , nixpkgs-2405
    } @inputs:
    flake-utils.lib.eachDefaultSystem
      (system:
      let
        pkgs-unstable = nixpkgs-unstable.legacyPackages.${system};
        pkgs-2505 = nixpkgs-2505.legacyPackages.${system};
        pkgs-2405 = nixpkgs-2405.legacyPackages.${system};
        atomi = atomipkgs.packages.${system};
        pre-commit-lib = pre-commit-hooks.lib.${system};
      in
      let pkgs = pkgs-2505; in
      let
        out = rec {
          pre-commit = import ./nix/pre-commit.nix {
            inherit pre-commit-lib formatter packages;
          };
          formatter = import ./nix/fmt.nix {
            inherit treefmt-nix pkgs;
          };
          packages = import ./nix/packages.nix {
            inherit pkgs atomi pkgs-2505 pkgs-2405 pkgs-unstable;
          };
          env = import ./nix/env.nix {
            inherit pkgs packages;
          };
          devShells = import ./nix/shells.nix {
            inherit pkgs env packages;
            shellHook = checks.pre-commit-check.shellHook;
          };
          checks = {
            pre-commit-check = pre-commit;
            format = formatter;
          };
        };
      in
      with out;
      {
        inherit checks formatter packages devShells;
      }
      );
}
