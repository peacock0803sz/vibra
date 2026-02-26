package server

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"connectrpc.com/connect"

	agentv1 "github.com/peacock0803sz/vibra/back/gen/vibra/agent/v1"
	"github.com/peacock0803sz/vibra/back/gen/vibra/agent/v1/agentv1connect"
	"github.com/peacock0803sz/vibra/back/internal/adapter"
	"github.com/peacock0803sz/vibra/back/internal/container"
	"github.com/peacock0803sz/vibra/back/internal/sandbox"
)

// fakeRuntime is a mock container runtime for testing.
type fakeRuntime struct {
	output string
}

func (f *fakeRuntime) Ping(context.Context) error                                     { return nil }
func (f *fakeRuntime) Create(context.Context, *adapter.ContainerSpec) (string, error)  { return "fake-id", nil }
func (f *fakeRuntime) Start(context.Context, string) error                             { return nil }
func (f *fakeRuntime) Kill(context.Context, string) error                              { return nil }
func (f *fakeRuntime) Remove(context.Context, string) error                            { return nil }
func (f *fakeRuntime) Attach(_ context.Context, _ string) (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader(f.output)), nil
}

func TestExecute_StreamsEvents(t *testing.T) {
	rt := &fakeRuntime{
		output: `{"type":"system","subtype":"init","session_id":"sess_test"}
{"type":"assistant","message":{"id":"msg_01","type":"text","text":"Hello from Claude"}}
{"type":"result","subtype":"success","session_id":"sess_test","cost_usd":0.001}
`,
	}
	runner := container.NewRunner(rt)
	cfg := &sandbox.Config{
		AllowedDirs: []string{"/tmp"},
		AllowedEnvs: []string{"ANTHROPIC_API_KEY"},
	}

	srv := NewAgentServer(runner, cfg, "")

	_, handler := agentv1connect.NewAgentServiceHandler(srv)
	ts := httptest.NewServer(handler)
	defer ts.Close()

	client := agentv1connect.NewAgentServiceClient(http.DefaultClient, ts.URL)

	stream, err := client.Execute(context.Background(), connect.NewRequest(&agentv1.ExecuteRequest{
		Agent:            agentv1.AgentType_AGENT_TYPE_CLAUDE,
		Prompt:           "hello",
		WorkingDirectory: "/tmp",
	}))
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	var events []*agentv1.StreamEvent
	for stream.Receive() {
		events = append(events, stream.Msg().Event)
	}
	if err := stream.Err(); err != nil {
		t.Fatalf("stream error: %v", err)
	}

	// EnvironmentInfo (sequence=1) + init session + text + result session = 4 events
	if len(events) != 4 {
		t.Fatalf("got %d events, want 4", len(events))
	}

	// environment info event (first, sequence=1)
	if _, ok := events[0].Payload.(*agentv1.StreamEvent_Environment); !ok {
		t.Errorf("event[0] = %T, want Environment", events[0].Payload)
	}

	// init session event
	if _, ok := events[1].Payload.(*agentv1.StreamEvent_Session); !ok {
		t.Errorf("event[1] = %T, want Session", events[1].Payload)
	}

	// text event
	text, ok := events[2].Payload.(*agentv1.StreamEvent_Text)
	if !ok {
		t.Fatalf("event[2] = %T, want Text", events[2].Payload)
	}
	if text.Text != "Hello from Claude" {
		t.Errorf("text = %q, want Hello from Claude", text.Text)
	}

	// verify sequence numbers are consecutive
	for i, ev := range events {
		if ev.Sequence != int64(i+1) {
			t.Errorf("event[%d].Sequence = %d, want %d", i, ev.Sequence, i+1)
		}
		if ev.Timestamp == nil {
			t.Errorf("event[%d].Timestamp is nil", i)
		}
	}
}

func TestExecute_InvalidAgent(t *testing.T) {
	runner := container.NewRunner(&fakeRuntime{})
	cfg := &sandbox.Config{AllowedDirs: []string{"/tmp"}}
	srv := NewAgentServer(runner, cfg, "")

	_, handler := agentv1connect.NewAgentServiceHandler(srv)
	ts := httptest.NewServer(handler)
	defer ts.Close()

	client := agentv1connect.NewAgentServiceClient(http.DefaultClient, ts.URL)

	stream, err := client.Execute(context.Background(), connect.NewRequest(&agentv1.ExecuteRequest{
		Agent:            agentv1.AgentType_AGENT_TYPE_CODEX,
		Prompt:           "hello",
		WorkingDirectory: "/tmp",
	}))
	// In server streaming, errors may be returned from the stream
	if err != nil {
		if connect.CodeOf(err) != connect.CodeInvalidArgument {
			t.Errorf("code = %v, want InvalidArgument", connect.CodeOf(err))
		}
		return
	}
	// If a stream was returned, check for errors via Receive
	for stream.Receive() {
	}
	err = stream.Err()
	if err == nil {
		t.Fatal("expected error for unsupported agent")
	}
	if connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Errorf("code = %v, want InvalidArgument", connect.CodeOf(err))
	}
}

func TestExecute_DisallowedDirectory(t *testing.T) {
	runner := container.NewRunner(&fakeRuntime{})
	cfg := &sandbox.Config{AllowedDirs: []string{"/tmp"}}
	srv := NewAgentServer(runner, cfg, "")

	_, handler := agentv1connect.NewAgentServiceHandler(srv)
	ts := httptest.NewServer(handler)
	defer ts.Close()

	client := agentv1connect.NewAgentServiceClient(http.DefaultClient, ts.URL)

	stream, err := client.Execute(context.Background(), connect.NewRequest(&agentv1.ExecuteRequest{
		Agent:            agentv1.AgentType_AGENT_TYPE_CLAUDE,
		Prompt:           "hello",
		WorkingDirectory: "/etc/secrets",
	}))
	if err != nil {
		if connect.CodeOf(err) != connect.CodeInvalidArgument {
			t.Errorf("code = %v, want InvalidArgument", connect.CodeOf(err))
		}
		return
	}
	for stream.Receive() {
	}
	err = stream.Err()
	if err == nil {
		t.Fatal("expected error for disallowed directory")
	}
	if connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Errorf("code = %v, want InvalidArgument", connect.CodeOf(err))
	}
}

func TestExecute_SkipsUnparseableLines(t *testing.T) {
	rt := &fakeRuntime{
		output: `not json
{"type":"assistant","message":{"id":"msg_01","type":"text","text":"works"}}
also not json
`,
	}
	runner := container.NewRunner(rt)
	cfg := &sandbox.Config{AllowedDirs: []string{"/tmp"}}
	srv := NewAgentServer(runner, cfg, "")

	_, handler := agentv1connect.NewAgentServiceHandler(srv)
	ts := httptest.NewServer(handler)
	defer ts.Close()

	client := agentv1connect.NewAgentServiceClient(http.DefaultClient, ts.URL)

	stream, err := client.Execute(context.Background(), connect.NewRequest(&agentv1.ExecuteRequest{
		Agent:            agentv1.AgentType_AGENT_TYPE_CLAUDE,
		Prompt:           "hello",
		WorkingDirectory: "/tmp",
	}))
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	var count int
	for stream.Receive() {
		count++
	}
	// EnvironmentInfo (sequence=1) + text event = 2; unparseable lines are skipped
	if count != 2 {
		t.Errorf("got %d events, want 2 (unparseable lines should be skipped)", count)
	}
}
