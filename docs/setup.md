# Vibra Setup Guide

Deploy Vibra on your personal infrastructure for vibe coding from any device.

## Prerequisites

- **Tailscale**: Vibra authenticates via Tailscale identity headers. Install and join your tailnet.
- **Container runtime**: Docker or Podman (for running AI agent containers).
- **Node.js 22**: Required for the frontend SSR server.
- **GitHub PAT** (for npm install): GitHub Packages requires authentication. Create a token with `read:packages` scope.

## Installation

### Option A: Download Pre-built Binaries

Download the latest release from [GitHub Releases](https://github.com/peacock0803sz/vibra/releases).

**Linux (amd64)**:

```bash
curl -Lo vibra.tar.gz https://github.com/peacock0803sz/vibra/releases/latest/download/vibra_linux_amd64.tar.gz
tar xzf vibra.tar.gz
sudo mv vibra /usr/local/bin/
```

**macOS (Apple Silicon)**:

```bash
curl -Lo vibra.tar.gz https://github.com/peacock0803sz/vibra/releases/latest/download/vibra_darwin_arm64.tar.gz
tar xzf vibra.tar.gz
sudo mv vibra /usr/local/bin/
```

### Option B: Install Frontend via npm

Configure your `.npmrc` for GitHub Packages:

```bash
echo "@peacock0803sz:registry=https://npm.pkg.github.com" >> ~/.npmrc
echo "//npm.pkg.github.com/:_authToken=YOUR_GITHUB_PAT" >> ~/.npmrc
```

Then install:

```bash
npm install -g @peacock0803sz/vibra
```

### Option C: NixOS / nix-darwin (Recommended)

Add the flake input and import the module:

**NixOS** (`/etc/nixos/flake.nix`):

```nix
{
  inputs.vibra.url = "github:peacock0803sz/vibra";

  outputs = { nixpkgs, vibra, ... }: {
    nixosConfigurations.myhost = nixpkgs.lib.nixosSystem {
      modules = [
        vibra.nixosModules.vibra
        {
          services.vibra = {
            enable = true;
            allowedDirs = [ "/home/you/projects" ];
            environmentFile = "/run/secrets/vibra.env";
          };
        }
      ];
    };
  };
}
```

**nix-darwin** (`~/.config/nix-darwin/flake.nix`):

```nix
{
  inputs.vibra.url = "github:peacock0803sz/vibra";

  outputs = { nix-darwin, vibra, ... }: {
    darwinConfigurations.myhost = nix-darwin.lib.darwinSystem {
      modules = [
        vibra.darwinModules.vibra
        {
          services.vibra = {
            enable = true;
            allowedDirs = [ "/Users/you/projects" ];
          };
        }
      ];
    };
  };
}
```

## Configuration

Vibra is configured via environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `VIBRA_LISTEN_ADDR` | `127.0.0.1:3001` | Backend listen address |
| `VIBRA_CORS_ORIGIN` | `http://127.0.0.1:5173` | Allowed CORS origin |
| `VIBRA_ALLOWED_DIRS` | (required) | Comma-separated sandbox directories |
| `VIBRA_ALLOWED_ENVS` | | Comma-separated env vars for agent containers |
| `VIBRA_DEV_USER` | | Dev-mode auth bypass username |
| `VIBRA_DEFAULT_WORKDIR` | | Default working directory |
| `ANTHROPIC_API_KEY` | | API key for Claude agent |
| `GOOGLE_API_KEY` | | API key for Gemini agent |
| `OPENAI_API_KEY` | | API key for Codex agent |

For NixOS deployments, use `environmentFile` to keep secrets out of the Nix store:

```bash
# /run/secrets/vibra.env
ANTHROPIC_API_KEY=sk-ant-...
GOOGLE_API_KEY=AI...
OPENAI_API_KEY=sk-...
```

## First Run

### Manual Start

```bash
# Terminal 1: Backend
export VIBRA_ALLOWED_DIRS="/home/you/projects"
export ANTHROPIC_API_KEY="sk-ant-..."
vibra

# Terminal 2: Frontend
npx @peacock0803sz/vibra
# or: vibra-front (if installed via Nix)
```

### Verify

Open `http://127.0.0.1:3000` in your browser. The Vibra UI should load and connect to the backend.

Check version:

```bash
vibra --version
```

## Service Management

### NixOS

```bash
sudo systemctl status vibra-back vibra-front
sudo systemctl restart vibra-back
sudo journalctl -u vibra-back -f
```

### nix-darwin

```bash
launchctl list | grep vibra
launchctl kickstart -k gui/$(id -u)/com.peacock0803sz.vibra-back
tail -f /tmp/vibra-back.log
```

## Upgrading

### Binary

Download the new release and replace the binary. Restart the service.

### npm

```bash
npm update -g @peacock0803sz/vibra
```

### NixOS / nix-darwin

Update the flake input and rebuild:

```bash
nix flake update vibra
sudo nixos-rebuild switch  # NixOS
darwin-rebuild switch      # nix-darwin
```

## Troubleshooting

### Container runtime not found

Vibra requires Docker or Podman to run agent containers. Verify:

```bash
docker info    # or: podman info
```

### Tailscale not connected

The backend reads Tailscale identity from HTTP headers. Ensure Tailscale is running:

```bash
tailscale status
```

For local development, set `VIBRA_DEV_USER` to bypass Tailscale auth.

### VIBRA_ALLOWED_DIRS is required

The backend requires at least one sandbox directory. Set it in your environment or Nix module config.

### npm install authentication error

GitHub Packages requires a PAT with `read:packages` scope. Verify your `.npmrc`:

```bash
npm config get @peacock0803sz:registry
# Should output: https://npm.pkg.github.com
```
