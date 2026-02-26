package container

import (
	agentv1 "github.com/peacock0803sz/vibra/back/gen/vibra/agent/v1"
	"github.com/peacock0803sz/vibra/back/internal/adapter"
)

const (
	// NetworkBridge is the default Docker bridge network (outbound internet, no host localhost).
	NetworkBridge = "bridge"
	// NetworkHost shares the host's network namespace (localhost accessible).
	NetworkHost = "host"
	// NetworkNone disables networking entirely.
	NetworkNone = "none"
)

// ApplyPermission configures ContainerSpec volumes and network based on the permission mode.
// UNSPECIFIED defaults to VERIFY.
func ApplyPermission(spec *adapter.ContainerSpec, mode agentv1.PermissionMode) {
	if mode == agentv1.PermissionMode_PERMISSION_MODE_UNSPECIFIED {
		mode = agentv1.PermissionMode_PERMISSION_MODE_VERIFY
	}

	switch mode {
	case agentv1.PermissionMode_PERMISSION_MODE_OBSERVE:
		// Read-only volumes + API-only network (bridge: outbound ok, no host localhost)
		setVolumesReadOnly(spec, true)
		spec.NetworkMode = NetworkBridge

	case agentv1.PermissionMode_PERMISSION_MODE_VERIFY:
		// Read-only volumes + API + localhost (for test execution, Playwright, etc.)
		setVolumesReadOnly(spec, true)
		spec.NetworkMode = NetworkHost

	case agentv1.PermissionMode_PERMISSION_MODE_EDIT:
		// Read-write volumes + API + localhost
		setVolumesReadOnly(spec, false)
		spec.NetworkMode = NetworkHost

	case agentv1.PermissionMode_PERMISSION_MODE_FULL:
		// Read-write volumes + unrestricted network
		setVolumesReadOnly(spec, false)
		spec.NetworkMode = NetworkHost
	}
}

func setVolumesReadOnly(spec *adapter.ContainerSpec, readOnly bool) {
	for i := range spec.Volumes {
		if !spec.Volumes[i].System {
			spec.Volumes[i].ReadOnly = readOnly
		}
	}
}
