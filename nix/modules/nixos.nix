{ config, lib, pkgs, ... }:
let
  cfg = config.services.vibra;
in
{
  options.services.vibra = {
    enable = lib.mkEnableOption "vibra service";

    backend = {
      enable = lib.mkOption {
        type = lib.types.bool;
        default = true;
        description = "Enable vibra backend service";
      };

      package = lib.mkOption {
        type = lib.types.package;
        description = "Vibra backend package";
      };

      listenAddr = lib.mkOption {
        type = lib.types.str;
        default = "127.0.0.1:3001";
        description = "Backend listen address";
      };

      corsOrigin = lib.mkOption {
        type = lib.types.str;
        default = "http://127.0.0.1:3000";
        description = "Allowed CORS origin";
      };

      allowedDirs = lib.mkOption {
        type = lib.types.listOf lib.types.str;
        description = "Sandbox working directories";
      };

      allowedEnvs = lib.mkOption {
        type = lib.types.listOf lib.types.str;
        default = [ "ANTHROPIC_API_KEY" "GOOGLE_API_KEY" "OPENAI_API_KEY" ];
        description = "Env vars passed to agent containers";
      };

      devUser = lib.mkOption {
        type = lib.types.nullOr lib.types.str;
        default = null;
        description = "Dev auth bypass user";
      };

      defaultWorkdir = lib.mkOption {
        type = lib.types.nullOr lib.types.str;
        default = null;
        description = "Default working directory";
      };

      environmentFile = lib.mkOption {
        type = lib.types.nullOr lib.types.path;
        default = null;
        description = "Secrets file path (systemd EnvironmentFile)";
      };
    };

    frontend = {
      enable = lib.mkOption {
        type = lib.types.bool;
        default = true;
        description = "Enable vibra frontend SSR service";
      };

      package = lib.mkOption {
        type = lib.types.package;
        description = "Vibra frontend package";
      };

      port = lib.mkOption {
        type = lib.types.port;
        default = 3000;
        description = "Frontend SSR server port";
      };
    };

    user = lib.mkOption {
      type = lib.types.str;
      default = "vibra";
      description = "Service user";
    };

    group = lib.mkOption {
      type = lib.types.str;
      default = "vibra";
      description = "Service group";
    };
  };

  config = lib.mkIf cfg.enable {
    users.users.${cfg.user} = {
      isSystemUser = true;
      group = cfg.group;
    };
    users.groups.${cfg.group} = { };

    systemd.services.vibra-back = lib.mkIf cfg.backend.enable {
      description = "Vibra backend server";
      after = [ "network.target" ];
      wantedBy = [ "multi-user.target" ];

      environment = {
        VIBRA_LISTEN_ADDR = cfg.backend.listenAddr;
        VIBRA_CORS_ORIGIN = cfg.backend.corsOrigin;
        VIBRA_ALLOWED_DIRS = lib.concatStringsSep "," cfg.backend.allowedDirs;
        VIBRA_ALLOWED_ENVS = lib.concatStringsSep "," cfg.backend.allowedEnvs;
      } // lib.optionalAttrs (cfg.backend.devUser != null) {
        VIBRA_DEV_USER = cfg.backend.devUser;
      } // lib.optionalAttrs (cfg.backend.defaultWorkdir != null) {
        VIBRA_DEFAULT_WORKDIR = cfg.backend.defaultWorkdir;
      };

      serviceConfig = {
        Type = "simple";
        ExecStart = "${cfg.backend.package}/bin/vibra";
        User = cfg.user;
        Group = cfg.group;
        Restart = "on-failure";
        RestartSec = 5;

        # Security hardening
        NoNewPrivileges = true;
        PrivateTmp = true;
        ProtectSystem = "strict";
        ProtectHome = true;
        ReadOnlyPaths = [ "/" ];
        ReadWritePaths = cfg.backend.allowedDirs ++ [ "/var/run/docker.sock" ];
      } // lib.optionalAttrs (cfg.backend.environmentFile != null) {
        EnvironmentFile = cfg.backend.environmentFile;
      };
    };

    systemd.services.vibra-front = lib.mkIf cfg.frontend.enable {
      description = "Vibra frontend SSR server";
      after = [ "network.target" ] ++ lib.optionals cfg.backend.enable [ "vibra-back.service" ];
      wantedBy = [ "multi-user.target" ];

      environment = {
        PORT = toString cfg.frontend.port;
      };

      serviceConfig = {
        Type = "simple";
        ExecStart = "${cfg.frontend.package}/bin/vibra-front";
        User = cfg.user;
        Group = cfg.group;
        Restart = "on-failure";
        RestartSec = 5;

        NoNewPrivileges = true;
        PrivateTmp = true;
        ProtectSystem = "strict";
        ProtectHome = true;
      };
    };
  };
}
