package server

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
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

	runner   *container.Runner
	sandbox  *sandbox.Config
	adapters map[agentv1.AgentType]adapter.AgentAdapter
}

// NewAgentServer creates an AgentServer with the given container runner and sandbox config.
func NewAgentServer(runner *container.Runner, sandbox *sandbox.Config) *AgentServer {
	return &AgentServer{
		runner:  runner,
		sandbox: sandbox,
		adapters: map[agentv1.AgentType]adapter.AgentAdapter{
			agentv1.AgentType_AGENT_TYPE_CLAUDE: adapter.NewClaudeAdapter(),
		},
	}
}

// Execute はプロンプトをAIエージェントに送信し、レスポンスをストリーミングする。
func (s *AgentServer) Execute(
	ctx context.Context,
	req *connect.Request[agentv1.ExecuteRequest],
	stream *connect.ServerStream[agentv1.ExecuteResponse],
) error {
	msg := req.Msg

	// アダプタ解決
	a, ok := s.adapters[msg.Agent]
	if !ok {
		return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("unknown agent type: %v", msg.Agent))
	}

	// ワーキングディレクトリ検証
	if err := s.sandbox.ValidateDir(msg.WorkingDirectory); err != nil {
		return connect.NewError(connect.CodeInvalidArgument, err)
	}

	// 環境変数フィルタリング
	filteredEnv := s.sandbox.FilterEnv(msg.Env)

	// ContainerSpec構築
	spec, err := a.Start(ctx, msg)
	if err != nil {
		return connect.NewError(connect.CodeInternal, fmt.Errorf("adapter start: %w", err))
	}
	spec.Env = append(spec.Env, filteredEnv...)

	// パーミッション適用
	container.ApplyPermission(spec, msg.PermissionMode)

	// コンテナ起動
	logs, err := s.runner.Run(ctx, spec)
	if err != nil {
		return connect.NewError(connect.CodeUnavailable, fmt.Errorf("container run: %w", err))
	}
	defer logs.Close()

	// stdout行ごとにパースしてストリーミング
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
	// 大きなJSON行に対応 (最大1MB)
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
