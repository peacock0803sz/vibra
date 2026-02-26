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
        default = "127.0.0.1:13001";
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
        description = ''
          Dev auth bypass user. WARNING: Setting this option bypasses all
          authentication and treats every request as coming from the specified
          user. Never enable this in production deployments.
        '';
      };

      defaultWorkdir = lib.mkOption {
        type = lib.types.nullOr lib.types.str;
        default = null;
        description = "Default working directory";
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
  };

  config = lib.mkIf cfg.enable {
    warnings = lib.optionals (cfg.backend.devUser != null) [
      "services.vibra.backend.devUser is set to \"${cfg.backend.devUser}\". This bypasses all authentication. Do not use in production."
    ];

    launchd.agents.vibra-back = lib.mkIf cfg.backend.enable {
      serviceConfig = {
        Label = "com.peacock0803sz.vibra-back";
        ProgramArguments = [ "${cfg.backend.package}/bin/vibra" ];
        KeepAlive = true;
        RunAtLoad = true;
        EnvironmentVariables = {
          VIBRA_LISTEN_ADDR = cfg.backend.listenAddr;
          VIBRA_CORS_ORIGIN = cfg.backend.corsOrigin;
          VIBRA_ALLOWED_DIRS = lib.concatStringsSep "," cfg.backend.allowedDirs;
          VIBRA_ALLOWED_ENVS = lib.concatStringsSep "," cfg.backend.allowedEnvs;
        } // lib.optionalAttrs (cfg.backend.devUser != null) {
          VIBRA_DEV_USER = cfg.backend.devUser;
        } // lib.optionalAttrs (cfg.backend.defaultWorkdir != null) {
          VIBRA_DEFAULT_WORKDIR = cfg.backend.defaultWorkdir;
        };
        StandardOutPath = "/tmp/vibra-back.log";
        StandardErrorPath = "/tmp/vibra-back.err";
        # NOTE: launchd has no EnvironmentFile equivalent.
        # API keys must be set in EnvironmentVariables directly
        # or loaded via a wrapper script.
      };
    };

    launchd.agents.vibra-front = lib.mkIf cfg.frontend.enable {
      serviceConfig = {
        Label = "com.peacock0803sz.vibra-front";
        ProgramArguments = [ "${cfg.frontend.package}/bin/vibra-front" ];
        KeepAlive = true;
        RunAtLoad = true;
        EnvironmentVariables = {
          PORT = toString cfg.frontend.port;
        };
        StandardOutPath = "/tmp/vibra-front.log";
        StandardErrorPath = "/tmp/vibra-front.err";
      };
    };
  };
}
