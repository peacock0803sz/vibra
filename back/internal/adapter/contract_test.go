package adapter

import (
	"bufio"
	"os"
	"testing"

	agentv1 "github.com/peacock0803sz/vibra/back/gen/vibra/agent/v1"
)

// TestClaudeAdapter_Contract is a contract test using snapshot files.
// It detects changes in the CLI output format early.
func TestClaudeAdapter_Contract(t *testing.T) {
	a := NewClaudeAdapter()

	f, err := os.Open("testdata/claude_stream_v1.jsonl")
	if err != nil {
		t.Fatalf("open snapshot: %v", err)
	}
	defer f.Close()

	var events []*agentv1.StreamEvent
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		if ev := a.ParseLine(line); ev != nil {
			events = append(events, ev)
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("scan: %v", err)
	}

	if len(events) == 0 {
		t.Fatal("no events parsed from snapshot")
	}

	// init event: SessionInfo
	assertPayloadType[*agentv1.StreamEvent_Session](t, events[0], "event[0] should be session init")

	// text event
	assertPayloadType[*agentv1.StreamEvent_Text](t, events[1], "event[1] should be text")

	// tool_use event
	assertPayloadType[*agentv1.StreamEvent_ToolUse](t, events[2], "event[2] should be tool_use")

	// tool_result event
	assertPayloadType[*agentv1.StreamEvent_ToolResult](t, events[3], "event[3] should be tool_result")

	// second text event
	assertPayloadType[*agentv1.StreamEvent_Text](t, events[4], "event[4] should be text")

	// result event: SessionInfo (with cost)
	sess := assertPayloadType[*agentv1.StreamEvent_Session](t, events[5], "event[5] should be session result")
	if sess.Session.CostMicros == 0 {
		t.Error("result event should have non-zero cost_micros")
	}
}

func assertPayloadType[T any](t *testing.T, ev *agentv1.StreamEvent, msg string) T {
	t.Helper()
	v, ok := ev.Payload.(T)
	if !ok {
		t.Fatalf("%s: got %T", msg, ev.Payload)
	}
	return v
}
