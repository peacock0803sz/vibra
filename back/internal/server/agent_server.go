package server

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"sync/atomic"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"

	agentv1 "github.com/peacock0803sz/vibra/back/gen/vibra/agent/v1"
	"github.com/peacock0803sz/vibra/back/gen/vibra/agent/v1/agentv1connect"
	"github.com/peacock0803sz/vibra/back/internal/adapter"
	"github.com/peacock0803sz/vibra/back/internal/container"
	"github.com/peacock0803sz/vibra/back/internal/sandbox"
)

// AgentServer implements agentv1connect.AgentServiceHandler.
type AgentServer struct {
	agentv1connect.UnimplementedAgentServiceHandler

	runner         *container.Runner
	sandbox        *sandbox.Config
	defaultWorkDir string
	adapters       map[agentv1.AgentType]adapter.AgentAdapter
}

// NewAgentServer creates an AgentServer with the given container runner and sandbox config.
// defaultWorkDir is used when the client does not specify a working directory.
func NewAgentServer(runner *container.Runner, sandbox *sandbox.Config, defaultWorkDir string) *AgentServer {
	return &AgentServer{
		runner:         runner,
		sandbox:        sandbox,
		defaultWorkDir: defaultWorkDir,
		adapters: map[agentv1.AgentType]adapter.AgentAdapter{
			agentv1.AgentType_AGENT_TYPE_CLAUDE: adapter.NewClaudeAdapter(),
		},
	}
}

// GetNodeInfo returns information about this vibra node.
func (s *AgentServer) GetNodeInfo(
	ctx context.Context,
	req *connect.Request[agentv1.GetNodeInfoRequest],
) (*connect.Response[agentv1.GetNodeInfoResponse], error) {
	hostname, err := os.Hostname()
	if err != nil {
		log.Printf("WARNING: os.Hostname() failed: %v", err)
		hostname = "unknown"
	}

	agents := s.availableAgents()

	return connect.NewResponse(&agentv1.GetNodeInfoResponse{
		NodeId:            hostname,
		TailscaleHostname: hostname,
		AvailableAgents:   agents,
		Status:            agentv1.NodeStatus_NODE_STATUS_ONLINE,
	}), nil
}

// ListAgents returns the agents available on this node.
func (s *AgentServer) ListAgents(
	ctx context.Context,
	req *connect.Request[agentv1.ListAgentsRequest],
) (*connect.Response[agentv1.ListAgentsResponse], error) {
	return connect.NewResponse(&agentv1.ListAgentsResponse{
		Agents: s.availableAgents(),
	}), nil
}

// availableAgents builds the AgentInfo list from registered adapters.
func (s *AgentServer) availableAgents() []*agentv1.AgentInfo {
	agents := make([]*agentv1.AgentInfo, 0, len(s.adapters))
	for agentType, a := range s.adapters {
		agents = append(agents, &agentv1.AgentInfo{
			Type:      agentType,
			Available: a.Available(),
		})
	}
	sort.Slice(agents, func(i, j int) bool {
		return agents[i].Type < agents[j].Type
	})
	return agents
}

// Execute sends a prompt to the AI agent and streams the response.
func (s *AgentServer) Execute(
	ctx context.Context,
	req *connect.Request[agentv1.ExecuteRequest],
	stream *connect.ServerStream[agentv1.ExecuteResponse],
) error {
	msg := req.Msg

	// Resolve adapter
	a, ok := s.adapters[msg.Agent]
	if !ok {
		return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("unknown agent type: %v", msg.Agent))
	}

	// Apply default working directory if not specified.
	if msg.WorkingDirectory == "" {
		if s.defaultWorkDir == "" {
			return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("working directory is required"))
		}
		msg.WorkingDirectory = s.defaultWorkDir
	}

	// Validate working directory
	if err := s.sandbox.ValidateDir(msg.WorkingDirectory); err != nil {
		return connect.NewError(connect.CodeInvalidArgument, err)
	}

	// Collect environment variables: server-side first, then client overrides.
	serverEnv := s.sandbox.CollectHostEnv()
	clientEnv := s.sandbox.FilterEnv(msg.Env)

	// Build ContainerSpec
	spec, err := a.Start(ctx, msg)
	if err != nil {
		return connect.NewError(connect.CodeInternal, fmt.Errorf("adapter start: %w", err))
	}
	spec.Env = append(spec.Env, serverEnv...)
	spec.Env = append(spec.Env, clientEnv...)

	// Apply permissions
	container.ApplyPermission(spec, msg.PermissionMode)

	// Start container
	logs, err := s.runner.Run(ctx, spec)
	if err != nil {
		return connect.NewError(connect.CodeUnavailable, fmt.Errorf("container run: %w", err))
	}
	defer logs.Close()

	// Parse stdout line by line and stream events
	return s.streamLines(ctx, a, logs, stream)
}

func (s *AgentServer) streamLines(
	ctx context.Context,
	a adapter.AgentAdapter,
	logs io.Reader,
	stream *connect.ServerStream[agentv1.ExecuteResponse],
) error {
	var seq atomic.Int64
	scanner := bufio.NewScanner(logs)
	// Handle large JSON lines (up to 1MB)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	for scanner.Scan() {
		if ctx.Err() != nil {
			return nil
		}

		line := scanner.Text()
		if line == "" {
			continue
		}

		ev := a.ParseLine(line)
		if ev == nil {
			log.Printf("skipping unparseable line: %.100s", line)
			continue
		}

		ev.Sequence = seq.Add(1)
		ev.Timestamp = timestamppb.Now()

		if err := stream.Send(&agentv1.ExecuteResponse{Event: ev}); err != nil {
			return err
		}
	}

	if err := scanner.Err(); err != nil && ctx.Err() == nil {
		return connect.NewError(connect.CodeInternal, fmt.Errorf("read container output: %w", err))
	}

	return nil
}
