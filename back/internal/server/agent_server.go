package server

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
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

	// Emit EnvironmentInfo as sequence=1 before container start (FR-001, FR-007)
	hostname, _ := os.Hostname()
	gitInfo := GetGitInfo(msg.WorkingDirectory)
	envEvent := &agentv1.StreamEvent{
		Sequence:  1,
		Timestamp: timestamppb.Now(),
		Payload: &agentv1.StreamEvent_Environment{
			Environment: &agentv1.EnvironmentInfo{
				Hostname:   hostname,
				Repository: gitInfo.Repository,
				Branch:     gitInfo.Branch,
				Agent:      msg.Agent,
				ModelName:  a.ModelName(),
			},
		},
	}
	if err := stream.Send(&agentv1.ExecuteResponse{Event: envEvent}); err != nil {
		return err
	}

	// Start container
	logs, err := s.runner.Run(ctx, spec)
	if err != nil {
		return connect.NewError(connect.CodeUnavailable, fmt.Errorf("container run: %w", err))
	}
	defer logs.Close()

	// Parse stdout line by line and stream events (sequence starts at 2)
	return s.streamLines(ctx, a, logs, stream)
}

func (s *AgentServer) streamLines(
	ctx context.Context,
	a adapter.AgentAdapter,
	logs io.Reader,
	stream *connect.ServerStream[agentv1.ExecuteResponse],
) error {
	var seq atomic.Int64
	seq.Store(1) // sequence=1 は EnvironmentInfo 用に予約済み
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
