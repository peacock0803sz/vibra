package server

import (
	"github.com/peacock0803sz/vibra/back/gen/vibra/agent/v1/agentv1connect"
	"github.com/peacock0803sz/vibra/back/internal/container"
	"github.com/peacock0803sz/vibra/back/internal/sandbox"
)

// AgentServer implements agentv1connect.AgentServiceHandler.
// Phase 2 provides the skeleton with UnimplementedAgentServiceHandler;
// the Execute streaming handler is implemented in Phase 3.
type AgentServer struct {
	agentv1connect.UnimplementedAgentServiceHandler

	runner  *container.Runner
	sandbox *sandbox.Config
}

// NewAgentServer creates an AgentServer with the given container runner and sandbox config.
func NewAgentServer(runner *container.Runner, sandbox *sandbox.Config) *AgentServer {
	return &AgentServer{
		runner:  runner,
		sandbox: sandbox,
	}
}
