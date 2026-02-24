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
