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
  };

  config = lib.mkIf cfg.enable {
    launchd.agents.vibra-back = {
      serviceConfig = {
        Label = "com.peacock0803sz.vibra-back";
        ProgramArguments = [ "${cfg.backPackage}/bin/vibra" ];
        KeepAlive = true;
        RunAtLoad = true;
        EnvironmentVariables = {
          VIBRA_LISTEN_ADDR = cfg.listenAddr;
          VIBRA_CORS_ORIGIN = cfg.corsOrigin;
          VIBRA_ALLOWED_DIRS = lib.concatStringsSep "," cfg.allowedDirs;
          VIBRA_ALLOWED_ENVS = lib.concatStringsSep "," cfg.allowedEnvs;
        } // lib.optionalAttrs (cfg.devUser != null) {
          VIBRA_DEV_USER = cfg.devUser;
        } // lib.optionalAttrs (cfg.defaultWorkdir != null) {
          VIBRA_DEFAULT_WORKDIR = cfg.defaultWorkdir;
        };
        StandardOutPath = "/tmp/vibra-back.log";
        StandardErrorPath = "/tmp/vibra-back.err";
        # NOTE: launchd has no EnvironmentFile equivalent.
        # API keys must be set in EnvironmentVariables directly
        # or loaded via a wrapper script.
      };
    };

    launchd.agents.vibra-front = {
      serviceConfig = {
        Label = "com.peacock0803sz.vibra-front";
        ProgramArguments = [ "${cfg.frontPackage}/bin/vibra-front" ];
        KeepAlive = true;
        RunAtLoad = true;
        EnvironmentVariables = {
          PORT = toString cfg.frontPort;
        };
        StandardOutPath = "/tmp/vibra-front.log";
        StandardErrorPath = "/tmp/vibra-front.err";
      };
    };
  };
}
