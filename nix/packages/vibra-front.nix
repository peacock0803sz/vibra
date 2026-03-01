{
  lib,
  stdenv,
  nodejs_24,
  pnpm,
  pnpmConfigHook,
  fetchPnpmDeps,
  makeWrapper,
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

  nativeBuildInputs = [ nodejs_24 pnpm pnpmConfigHook makeWrapper ];

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
    hash = if stdenv.hostPlatform.isLinux
      then "sha256-NVCXFbYOxxjRp7QXYOqXdfO1Ch3DKOCK3E6MPMrCtw8="
      else "sha256-M503XgV+834Yas2ofEgMG9kcuDfcQFlYudbwm5RjaKI=";
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

    # Prune devDependencies for production
    pnpm prune --prod --ignore-scripts
    find node_modules -xtype l -delete
    rm -f node_modules/.modules.yaml node_modules/.pnpm-workspace-state-v1.json

    mkdir -p "$out/lib/vibra-front"
    cp -r build "$out/lib/vibra-front/"
    cp -r node_modules "$out/lib/vibra-front/"
    cp start.js "$out/lib/vibra-front/"
    cp package.json "$out/lib/vibra-front/"

    makeWrapper ${nodejs_24}/bin/node "$out/bin/vibra-front" \
      --add-flags "$out/lib/vibra-front/start.js" \
      --set NODE_ENV production

    runHook postInstall
  '';

  meta = {
    description = "Vibra frontend SSR server (React Router)";
    mainProgram = "vibra-front";
  };
}
