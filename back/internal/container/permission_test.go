package container

import (
	"testing"

	agentv1 "github.com/peacock0803sz/vibra/back/gen/vibra/agent/v1"
	"github.com/peacock0803sz/vibra/back/internal/adapter"
)

func makeSpec() *adapter.ContainerSpec {
	return &adapter.ContainerSpec{
		Image: "test:latest",
		Volumes: []adapter.VolumeMount{
			{HostPath: "/src", ContainerPath: "/workspace"},
		},
	}
}

func TestApplyPermission_Observe(t *testing.T) {
	spec := makeSpec()
	ApplyPermission(spec, agentv1.PermissionMode_PERMISSION_MODE_OBSERVE)

	if !spec.Volumes[0].ReadOnly {
		t.Error("OBSERVE should set volumes to read-only")
	}
	if spec.NetworkMode != NetworkBridge {
		t.Errorf("OBSERVE network: got %q, want %q", spec.NetworkMode, NetworkBridge)
	}
}

func TestApplyPermission_Verify(t *testing.T) {
	spec := makeSpec()
	ApplyPermission(spec, agentv1.PermissionMode_PERMISSION_MODE_VERIFY)

	if !spec.Volumes[0].ReadOnly {
		t.Error("VERIFY should set volumes to read-only")
	}
	if spec.NetworkMode != NetworkHost {
		t.Errorf("VERIFY network: got %q, want %q", spec.NetworkMode, NetworkHost)
	}
}

func TestApplyPermission_Edit(t *testing.T) {
	spec := makeSpec()
	ApplyPermission(spec, agentv1.PermissionMode_PERMISSION_MODE_EDIT)

	if spec.Volumes[0].ReadOnly {
		t.Error("EDIT should set volumes to read-write")
	}
	if spec.NetworkMode != NetworkHost {
		t.Errorf("EDIT network: got %q, want %q", spec.NetworkMode, NetworkHost)
	}
}

func TestApplyPermission_Full(t *testing.T) {
	spec := makeSpec()
	ApplyPermission(spec, agentv1.PermissionMode_PERMISSION_MODE_FULL)

	if spec.Volumes[0].ReadOnly {
		t.Error("FULL should set volumes to read-write")
	}
	if spec.NetworkMode != NetworkHost {
		t.Errorf("FULL network: got %q, want %q", spec.NetworkMode, NetworkHost)
	}
}

func TestApplyPermission_SystemVolumePreserved(t *testing.T) {
	spec := &adapter.ContainerSpec{
		Image: "test:latest",
		Volumes: []adapter.VolumeMount{
			{HostPath: "/src", ContainerPath: "/workspace"},
			{HostPath: "named-vol", ContainerPath: "/home/agent", IsVolume: true, System: true},
		},
	}

	// OBSERVE sets non-system volumes to read-only but must not touch system volumes.
	ApplyPermission(spec, agentv1.PermissionMode_PERMISSION_MODE_OBSERVE)

	if !spec.Volumes[0].ReadOnly {
		t.Error("non-system volume should be read-only in OBSERVE mode")
	}
	if spec.Volumes[1].ReadOnly {
		t.Error("system volume should remain read-write regardless of permission mode")
	}

	// EDIT sets non-system volumes to read-write; system volume should still be unaffected.
	ApplyPermission(spec, agentv1.PermissionMode_PERMISSION_MODE_EDIT)

	if spec.Volumes[0].ReadOnly {
		t.Error("non-system volume should be read-write in EDIT mode")
	}
	if spec.Volumes[1].ReadOnly {
		t.Error("system volume should remain read-write regardless of permission mode")
	}
}

func TestApplyPermission_Unspecified(t *testing.T) {
	spec := makeSpec()
	ApplyPermission(spec, agentv1.PermissionMode_PERMISSION_MODE_UNSPECIFIED)

	// UNSPECIFIED defaults to VERIFY behavior.
	if !spec.Volumes[0].ReadOnly {
		t.Error("UNSPECIFIED (defaulting to VERIFY) should set volumes to read-only")
	}
	if spec.NetworkMode != NetworkHost {
		t.Errorf("UNSPECIFIED network: got %q, want %q", spec.NetworkMode, NetworkHost)
	}
}
