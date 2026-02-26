{ config, lib, pkgs, ... }:
let
  cfg = config.services.vibra;
in
{
  options.services.vibra = {
    enable = lib.mkEnableOption "vibra service";

    backPackage = lib.mkOption {
      type = lib.types.package;
      description = "Vibra backend package";
    };

    frontPackage = lib.mkOption {
      type = lib.types.package;
      description = "Vibra frontend package";
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

    frontPort = lib.mkOption {
      type = lib.types.port;
      default = 3000;
      description = "Frontend SSR server port";
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

    environmentFile = lib.mkOption {
      type = lib.types.nullOr lib.types.path;
      default = null;
      description = "Secrets file path (systemd EnvironmentFile)";
    };
  };

  config = lib.mkIf cfg.enable {
    users.users.${cfg.user} = {
      isSystemUser = true;
      group = cfg.group;
    };
    users.groups.${cfg.group} = { };

    systemd.services.vibra-back = {
      description = "Vibra backend server";
      after = [ "network.target" ];
      wantedBy = [ "multi-user.target" ];

      environment = {
        VIBRA_LISTEN_ADDR = cfg.listenAddr;
        VIBRA_CORS_ORIGIN = cfg.corsOrigin;
        VIBRA_ALLOWED_DIRS = lib.concatStringsSep "," cfg.allowedDirs;
        VIBRA_ALLOWED_ENVS = lib.concatStringsSep "," cfg.allowedEnvs;
      } // lib.optionalAttrs (cfg.devUser != null) {
        VIBRA_DEV_USER = cfg.devUser;
      } // lib.optionalAttrs (cfg.defaultWorkdir != null) {
        VIBRA_DEFAULT_WORKDIR = cfg.defaultWorkdir;
      };

      serviceConfig = {
        Type = "simple";
        ExecStart = "${cfg.backPackage}/bin/vibra";
        User = cfg.user;
        Group = cfg.group;
        Restart = "on-failure";
        RestartSec = 5;

        # セキュリティ強化
        NoNewPrivileges = true;
        PrivateTmp = true;
        ProtectSystem = "strict";
        ProtectHome = true;
        ReadOnlyPaths = [ "/" ];
        ReadWritePaths = cfg.allowedDirs;
      } // lib.optionalAttrs (cfg.environmentFile != null) {
        EnvironmentFile = cfg.environmentFile;
      };
    };

    systemd.services.vibra-front = {
      description = "Vibra frontend SSR server";
      after = [ "network.target" "vibra-back.service" ];
      wantedBy = [ "multi-user.target" ];

      environment = {
        PORT = toString cfg.frontPort;
      };

      serviceConfig = {
        Type = "simple";
        ExecStart = "${cfg.frontPackage}/bin/vibra-front";
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
