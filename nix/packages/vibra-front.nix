{
  lib,
  stdenv,
  nodejs_22,
  pnpm,
  pnpmConfigHook,
  fetchPnpmDeps,
}:
stdenv.mkDerivation {
  pname = "vibra-front";
  version = "0.0.0-dev";

  src = lib.fileset.toSource {
    root = ../..;
    fileset = lib.fileset.unions [
      ../../front
    ];
  };

  nativeBuildInputs = [ nodejs_22 pnpm pnpmConfigHook ];

  pnpmDeps = fetchPnpmDeps {
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
    hash = "sha256-Ghe/a12HQmWn9kYpaWlBtw4/P+2rmY4nlFKL7sVSmI4=";
    fetcherVersion = 3;
  };

  sourceRoot = "source/front";

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
    #!/usr/bin/env bash
    exec ${nodejs_22}/bin/node "$out/lib/vibra-front/start.js" "\$@"
    WRAPPER
    chmod +x "$out/bin/vibra-front"
    runHook postInstall
  '';

  meta = {
    description = "Vibra frontend SSR server (React Router)";
    mainProgram = "vibra-front";
  };
}
