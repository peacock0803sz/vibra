package adapter

import (
	"context"
	"encoding/json"
	"os/exec"

	agentv1 "github.com/peacock0803sz/vibra/back/gen/vibra/agent/v1"
)

const claudeImage = "vibra-claude"

// ClaudeAdapter implements AgentAdapter for Claude Code CLI.
type ClaudeAdapter struct{}

func NewClaudeAdapter() *ClaudeAdapter {
	return &ClaudeAdapter{}
}

func (a *ClaudeAdapter) Start(_ context.Context, req *agentv1.ExecuteRequest) (*ContainerSpec, error) {
	cmd := []string{
		"claude", "-p", req.Prompt,
		"--output-format", "stream-json",
	}

	if req.SessionId != "" {
		cmd = append(cmd, "--resume", req.SessionId)
	}

	return &ContainerSpec{
		Image:   claudeImage,
		Command: cmd,
		WorkDir: "/workspace",
		Volumes: []VolumeMount{
			{
				HostPath:      req.WorkingDirectory,
				ContainerPath: "/workspace",
			},
		},
		CPUQuota:    200000, // 2 CPU cores
		MemoryLimit: 2 * 1024 * 1024 * 1024, // 2 GiB
	}, nil
}

// claudeEvent is the top-level structure for Claude Code stream-json output.
type claudeEvent struct {
	Type    string       `json:"type"`
	Subtype string       `json:"subtype,omitempty"`
	Message *claudeMsg   `json:"message,omitempty"`
	// Fields for result events
	SessionID string  `json:"session_id,omitempty"`
	CostUSD   float64 `json:"cost_usd,omitempty"`
}

type claudeMsg struct {
	ID      string `json:"id,omitempty"`
	Type    string `json:"type"` // "text", "tool_use", "tool_result"
	Text    string `json:"text,omitempty"`
	Name    string `json:"name,omitempty"`
	Content string `json:"content,omitempty"`
	IsError bool   `json:"is_error,omitempty"`
}

func (a *ClaudeAdapter) ParseLine(line string) *agentv1.StreamEvent {
	var ev claudeEvent
	if err := json.Unmarshal([]byte(line), &ev); err != nil {
		return nil
	}

	switch ev.Type {
	case "system":
		if ev.Subtype == "init" && ev.SessionID != "" {
			return &agentv1.StreamEvent{
				Payload: &agentv1.StreamEvent_Session{
					Session: &agentv1.SessionInfo{
						SessionId: ev.SessionID,
						Agent:     agentv1.AgentType_AGENT_TYPE_CLAUDE,
					},
				},
			}
		}
		return nil

	case "assistant":
		if ev.Message == nil {
			return nil
		}
		return a.parseMessage(ev.Message)

	case "result":
		if ev.SessionID != "" {
			costMicros := int64(ev.CostUSD * 1_000_000)
			return &agentv1.StreamEvent{
				Payload: &agentv1.StreamEvent_Session{
					Session: &agentv1.SessionInfo{
						SessionId:    ev.SessionID,
						Agent:        agentv1.AgentType_AGENT_TYPE_CLAUDE,
						CostMicros:   costMicros,
						CurrencyCode: "USD",
					},
				},
			}
		}
		return nil

	default:
		return nil
	}
}

func (a *ClaudeAdapter) parseMessage(msg *claudeMsg) *agentv1.StreamEvent {
	switch msg.Type {
	case "text":
		return &agentv1.StreamEvent{
			Payload: &agentv1.StreamEvent_Text{Text: msg.Text},
		}

	case "tool_use":
		return &agentv1.StreamEvent{
			Payload: &agentv1.StreamEvent_ToolUse{
				ToolUse: &agentv1.ToolUse{ToolName: msg.Name},
			},
		}

	case "tool_result":
		return &agentv1.StreamEvent{
			Payload: &agentv1.StreamEvent_ToolResult{
				ToolResult: &agentv1.ToolResult{
					ToolName: msg.Name,
					Success:  !msg.IsError,
					Output:   msg.Content,
				},
			},
		}

	default:
		return nil
	}
}

func (a *ClaudeAdapter) ContinueFlags(sessionID string) []string {
	if sessionID == "" {
		return nil
	}
	return []string{"--resume", sessionID}
}

func (a *ClaudeAdapter) Available() bool {
	return exec.Command("docker", "image", "inspect", claudeImage).Run() == nil
}
