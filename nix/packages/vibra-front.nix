{
  lib,
  stdenv,
  nodejs_22,
  pnpm_10,
  buf,
  protobuf,
}:
stdenv.mkDerivation {
  pname = "vibra-front";
  version = "0.0.0-dev";

  src = lib.fileset.toSource {
    root = ../..;
    fileset = lib.fileset.unions [
      ../../front
      ../../proto
      ../../buf.gen.yaml
      ../../buf.yaml
    ];
  };

  nativeBuildInputs = [ nodejs_22 pnpm_10.configHook buf protobuf ];

  pnpmDeps = pnpm_10.fetchDeps {
    inherit (stdenv) system;
    pname = "vibra-front-deps";
    version = "0.0.0-dev";
    src = lib.fileset.toSource {
      root = ../..;
      fileset = lib.fileset.unions [
        ../../front/package.json
        ../../front/pnpm-lock.yaml
      ];
    };
    sourceRoot = "source/front";
    hash = lib.fakeHash;
  };

  sourceRoot = "source/front";

  preBuild = ''
    (cd "$NIX_BUILD_TOP/source" && buf generate)
  '';

  buildPhase = ''
    runHook preBuild
    pnpm build
    runHook postBuild
  '';

  installPhase = ''
    runHook preInstall
    mkdir -p "$out/lib/vibra-front"
    cp -r build "$out/lib/vibra-front/"
    cp start.js "$out/lib/vibra-front/"
    cp package.json "$out/lib/vibra-front/"

    mkdir -p "$out/bin"
    cat > "$out/bin/vibra-front" <<WRAPPER
    #!${nodejs_22}/bin/node
    process.argv = [process.argv[0], "$out/lib/vibra-front/start.js"];
    await import("$out/lib/vibra-front/start.js");
    WRAPPER
    chmod +x "$out/bin/vibra-front"
    runHook postInstall
  '';

  meta = {
    description = "Vibra frontend SSR server (React Router)";
    mainProgram = "vibra-front";
  };
}
