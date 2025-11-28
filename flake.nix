{
  description = "A launcher tool for managing application entries";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in
      {
        packages = {
          default = pkgs.callPackage ./default.nix { };
          launchit = pkgs.callPackage ./default.nix { };
        };

        # Development shell with Go toolchain
        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            go
          ];
        };
      }
    ) // {
      # For NixOS modules
      nixosModules.default = { config, lib, pkgs, ... }: {
        options.programs.launchit = {
          enable = lib.mkEnableOption "launchit";
        };

        config = lib.mkIf config.programs.launchit.enable {
          environment.systemPackages = [
            (pkgs.callPackage ./default.nix { })
          ];
        };
      };
    };
}
