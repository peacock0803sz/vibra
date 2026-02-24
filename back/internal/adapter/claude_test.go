package adapter

import (
	"context"
	"testing"

	agentv1 "github.com/peacock0803sz/vibra/back/gen/vibra/agent/v1"
)

func TestClaudeAdapter_Start(t *testing.T) {
	a := NewClaudeAdapter()
	req := &agentv1.ExecuteRequest{
		Agent:            agentv1.AgentType_AGENT_TYPE_CLAUDE,
		Prompt:           "hello",
		WorkingDirectory: "/home/user/project",
	}

	spec, err := a.Start(context.Background(), req)
	if err != nil {
		t.Fatalf("Start: %v", err)
	}

	if spec.Image != "vibra-claude" {
		t.Errorf("Image = %q, want vibra-claude", spec.Image)
	}
	if len(spec.Command) < 5 {
		t.Fatalf("Command too short: %v", spec.Command)
	}
	if spec.Command[0] != "claude" || spec.Command[2] != "hello" {
		t.Errorf("Command = %v, want claude -p hello ...", spec.Command)
	}
	if spec.WorkDir != "/workspace" {
		t.Errorf("WorkDir = %q, want /workspace", spec.WorkDir)
	}
	if len(spec.Volumes) != 1 || spec.Volumes[0].HostPath != "/home/user/project" {
		t.Errorf("Volumes = %v, want host mount to /home/user/project", spec.Volumes)
	}
}

func TestClaudeAdapter_Start_WithSessionID(t *testing.T) {
	a := NewClaudeAdapter()
	req := &agentv1.ExecuteRequest{
		Agent:            agentv1.AgentType_AGENT_TYPE_CLAUDE,
		Prompt:           "continue",
		SessionId:        "sess_abc123",
		WorkingDirectory: "/tmp",
	}

	spec, err := a.Start(context.Background(), req)
	if err != nil {
		t.Fatalf("Start: %v", err)
	}

	// --resume sess_abc123 が含まれるか確認
	found := false
	for i, arg := range spec.Command {
		if arg == "--resume" && i+1 < len(spec.Command) && spec.Command[i+1] == "sess_abc123" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Command missing --resume sess_abc123: %v", spec.Command)
	}
}

func TestClaudeAdapter_ParseLine_Text(t *testing.T) {
	a := NewClaudeAdapter()
	line := `{"type":"assistant","message":{"id":"msg_01","type":"text","text":"Hello world"}}`

	ev := a.ParseLine(line)
	if ev == nil {
		t.Fatal("ParseLine returned nil for text event")
	}

	text, ok := ev.Payload.(*agentv1.StreamEvent_Text)
	if !ok {
		t.Fatalf("payload type = %T, want StreamEvent_Text", ev.Payload)
	}
	if text.Text != "Hello world" {
		t.Errorf("text = %q, want Hello world", text.Text)
	}
}

func TestClaudeAdapter_ParseLine_ToolUse(t *testing.T) {
	a := NewClaudeAdapter()
	line := `{"type":"assistant","message":{"id":"msg_02","type":"tool_use","name":"Read","input":{"file_path":"/workspace/main.go"}}}`

	ev := a.ParseLine(line)
	if ev == nil {
		t.Fatal("ParseLine returned nil for tool_use event")
	}

	tu, ok := ev.Payload.(*agentv1.StreamEvent_ToolUse)
	if !ok {
		t.Fatalf("payload type = %T, want StreamEvent_ToolUse", ev.Payload)
	}
	if tu.ToolUse.ToolName != "Read" {
		t.Errorf("tool_name = %q, want Read", tu.ToolUse.ToolName)
	}
}

func TestClaudeAdapter_ParseLine_ToolResult(t *testing.T) {
	a := NewClaudeAdapter()
	line := `{"type":"assistant","message":{"id":"msg_03","type":"tool_result","name":"Read","content":"package main","is_error":false}}`

	ev := a.ParseLine(line)
	if ev == nil {
		t.Fatal("ParseLine returned nil for tool_result event")
	}

	tr, ok := ev.Payload.(*agentv1.StreamEvent_ToolResult)
	if !ok {
		t.Fatalf("payload type = %T, want StreamEvent_ToolResult", ev.Payload)
	}
	if tr.ToolResult.ToolName != "Read" {
		t.Errorf("tool_name = %q, want Read", tr.ToolResult.ToolName)
	}
	if !tr.ToolResult.Success {
		t.Error("expected success=true")
	}
}

func TestClaudeAdapter_ParseLine_SystemInit(t *testing.T) {
	a := NewClaudeAdapter()
	line := `{"type":"system","subtype":"init","session_id":"sess_abc123","tools":["Read","Write"]}`

	ev := a.ParseLine(line)
	if ev == nil {
		t.Fatal("ParseLine returned nil for system init event")
	}

	sess, ok := ev.Payload.(*agentv1.StreamEvent_Session)
	if !ok {
		t.Fatalf("payload type = %T, want StreamEvent_Session", ev.Payload)
	}
	if sess.Session.SessionId != "sess_abc123" {
		t.Errorf("session_id = %q, want sess_abc123", sess.Session.SessionId)
	}
}

func TestClaudeAdapter_ParseLine_Result(t *testing.T) {
	a := NewClaudeAdapter()
	line := `{"type":"result","subtype":"success","session_id":"sess_abc123","cost_usd":0.003}`

	ev := a.ParseLine(line)
	if ev == nil {
		t.Fatal("ParseLine returned nil for result event")
	}

	sess, ok := ev.Payload.(*agentv1.StreamEvent_Session)
	if !ok {
		t.Fatalf("payload type = %T, want StreamEvent_Session", ev.Payload)
	}
	if sess.Session.CostMicros != 3000 {
		t.Errorf("cost_micros = %d, want 3000", sess.Session.CostMicros)
	}
}

func TestClaudeAdapter_ParseLine_InvalidJSON(t *testing.T) {
	a := NewClaudeAdapter()
	ev := a.ParseLine("not json at all")
	if ev != nil {
		t.Errorf("expected nil for invalid JSON, got %v", ev)
	}
}

func TestClaudeAdapter_ParseLine_UnknownType(t *testing.T) {
	a := NewClaudeAdapter()
	ev := a.ParseLine(`{"type":"unknown","data":{}}`)
	if ev != nil {
		t.Errorf("expected nil for unknown type, got %v", ev)
	}
}

func TestClaudeAdapter_ContinueFlags(t *testing.T) {
	a := NewClaudeAdapter()

	flags := a.ContinueFlags("")
	if len(flags) != 1 || flags[0] != "--continue" {
		t.Errorf("ContinueFlags(\"\") = %v, want [--continue]", flags)
	}

	flags = a.ContinueFlags("sess_123")
	if len(flags) != 2 || flags[0] != "--resume" || flags[1] != "sess_123" {
		t.Errorf("ContinueFlags(sess_123) = %v, want [--resume sess_123]", flags)
	}
}
