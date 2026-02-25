package adapter

import (
	"context"

	agentv1 "github.com/peacock0803sz/vibra/back/gen/vibra/agent/v1"
)

// VolumeMount represents a bind mount or named volume for a container.
type VolumeMount struct {
	HostPath      string // host path (bind) or volume name (volume)
	ContainerPath string
	ReadOnly      bool
	System        bool // if true, ApplyPermission won't override ReadOnly
	IsVolume      bool // if true, HostPath is treated as a Docker named volume
}

// ContainerSpec defines the configuration for running an agent container.
// Adapters build this spec; the container runner executes it.
type ContainerSpec struct {
	Image       string
	Command     []string
	Env         []string
	Volumes     []VolumeMount
	NetworkMode string
	CPUQuota    int64 // microseconds per CPU period (100000 = 1 core)
	MemoryLimit int64 // bytes
	WorkDir     string
}

// AgentAdapter defines the interface for CLI agent implementations.
// Each adapter encapsulates CLI-specific logic (flags, output format,
// session continuation) for a particular AI agent.
type AgentAdapter interface {
	// Start builds a ContainerSpec for executing the agent with the given request.
	// The caller applies permission_mode to the ContainerSpec separately.
	Start(ctx context.Context, req *agentv1.ExecuteRequest) (*ContainerSpec, error)

	// ParseLine parses a single line of stdout from the agent CLI.
	// Returns nil if the line should be skipped (e.g., non-JSON debug output).
	ParseLine(line string) *agentv1.StreamEvent

	// ContinueFlags returns the CLI flags needed to resume a previous session.
	ContinueFlags(sessionID string) []string

	// Available reports whether the agent CLI is installed and version-compatible.
	Available() bool
}
