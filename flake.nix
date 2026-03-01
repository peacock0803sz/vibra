{
  description = "vibra - Vibe coding from any device";

  inputs = {
    flake-parts.url = "github:hercules-ci/flake-parts";
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    probitas.url = "github:probitas-test/probitas";
  };

  outputs = inputs@{ flake-parts, ... }:
    flake-parts.lib.mkFlake { inherit inputs; } {
      imports = [
        ./nix/packages.nix
        ./nix/modules.nix
      ];
      systems = [ "x86_64-linux" "aarch64-linux" "aarch64-darwin" "x86_64-darwin" ];
      perSystem = { pkgs, inputs', ... }: {
        devShells.default = pkgs.mkShell {
          packages = with pkgs; [
            # Go
            go

            # Node.js
            nodejs_24
            corepack_24

            # Protobuf / Buf
            buf
            protobuf

            # Container runtime
            docker-client

            # E2E testing
            inputs'.probitas.packages.default
          ];
        };
      };
    };
}
