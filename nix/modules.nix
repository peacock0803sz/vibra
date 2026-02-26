# flake-parts module: NixOS + nix-darwin service modules
{ withSystem, ... }:
{
  flake = {
    nixosModules.vibra = { pkgs, lib, ... }:
      let
        packages = withSystem pkgs.stdenv.hostPlatform.system (
          { config, ... }: {
            inherit (config.packages) vibra-back vibra-front;
          }
        );
      in
      {
        imports = [ ./modules/nixos.nix ];
        config.services.vibra = {
          backend.package = lib.mkDefault packages.vibra-back;
          frontend.package = lib.mkDefault packages.vibra-front;
        };
      };

    darwinModules.vibra = { pkgs, lib, ... }:
      let
        packages = withSystem pkgs.stdenv.hostPlatform.system (
          { config, ... }: {
            inherit (config.packages) vibra-back vibra-front;
          }
        );
      in
      {
        imports = [ ./modules/darwin.nix ];
        config.services.vibra = {
          backend.package = lib.mkDefault packages.vibra-back;
          frontend.package = lib.mkDefault packages.vibra-front;
        };
      };
  };
}
