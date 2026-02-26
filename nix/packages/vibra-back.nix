{
  lib,
  buildGoModule,
}:
buildGoModule {
  pname = "vibra-back";
  version = "0.0.0-dev";

  src = lib.fileset.toSource {
    root = ../..;
    fileset = lib.fileset.unions [
      ../../back
    ];
  };

  modRoot = "back";
  subPackages = [ "cmd/vibra" ];
  vendorHash = "sha256-UUb3lqjkbAtjhMptdWKYxo81PgYyJEZpc7OLgZmPKh0=";

  env.CGO_ENABLED = 0;

  ldflags = [
    "-s"
    "-w"
    "-X main.version=0.0.0-dev"
    "-X main.commit=nix"
    "-X main.date=unknown"
  ];

  meta = {
    description = "Vibra backend server (connect-go)";
    mainProgram = "vibra";
  };
}
