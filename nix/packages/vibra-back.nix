{
  lib,
  buildGoModule,
  buf,
  protobuf,
}:
buildGoModule {
  pname = "vibra-back";
  version = "0.0.0-dev";

  src = lib.fileset.toSource {
    root = ../..;
    fileset = lib.fileset.unions [
      ../../back
      ../../proto
      ../../buf.gen.yaml
      ../../buf.yaml
    ];
  };

  modRoot = "back";
  subPackages = [ "cmd/vibra" ];
  vendorHash = null;

  env.CGO_ENABLED = 0;

  ldflags = [
    "-s"
    "-w"
    "-X main.version=${builtins.placeholder "version" or "0.0.0-dev"}"
  ];

  nativeBuildInputs = [ buf protobuf ];

  preBuild = ''
    (cd "$NIX_BUILD_TOP/$sourceRoot" && buf generate)
  '';

  meta = {
    description = "Vibra backend server (connect-go)";
    mainProgram = "vibra";
  };
}
