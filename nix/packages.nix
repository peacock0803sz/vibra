# flake-parts module: perSystem packages
{ inputs, ... }:
{
  perSystem = { pkgs, system, ... }: {
    packages = {
      vibra-back = pkgs.callPackage ./packages/vibra-back.nix { };
      vibra-front = pkgs.callPackage ./packages/vibra-front.nix { };
    };
  };
}
