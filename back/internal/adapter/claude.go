package adapter

import (
	"context"
	"encoding/json"
	"os/exec"
	"strings"

	agentv1 "github.com/peacock0803sz/vibra/back/gen/vibra/agent/v1"
)

const (
	claudeImage      = "vibra-claude"
	claudeHomeVolume = "vibra-claude-home"
)

// ClaudeAdapter implements AgentAdapter for Claude Code CLI.
type ClaudeAdapter struct{}

func NewClaudeAdapter() *ClaudeAdapter {
	return &ClaudeAdapter{}
}

func (a *ClaudeAdapter) Start(_ context.Context, req *agentv1.ExecuteRequest) (*ContainerSpec, error) {
	cmd := []string{
		"claude", "-p", req.Prompt,
		"--output-format", "stream-json",
		"--verbose",
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
			{
				HostPath:      claudeHomeVolume,
				ContainerPath: "/home/agent",
				IsVolume:      true,
				System:        true,
			},
		},
		CPUQuota:    200000,                  // 2 CPU cores — TODO: make configurable per adapter
		MemoryLimit: 2 * 1024 * 1024 * 1024, // 2 GiB — TODO: make configurable per adapter
	}, nil
}

// claudeEvent is the top-level structure for Claude Code stream-json output.
type claudeEvent struct {
	Type    string     `json:"type"`
	Subtype string     `json:"subtype,omitempty"`
	Message *claudeMsg `json:"message,omitempty"`
	// Fields for result events
	SessionID    string  `json:"session_id,omitempty"`
	CostUSD      float64 `json:"cost_usd,omitempty"`
	TotalCostUSD float64 `json:"total_cost_usd,omitempty"`
}

// claudeMsg represents both the simple and verbose message formats.
type claudeMsg struct {
	ID      string              `json:"id,omitempty"`
	Type    string              `json:"type"`
	Text    string              `json:"text,omitempty"`
	Name    string              `json:"name,omitempty"`
	Content json.RawMessage     `json:"content,omitempty"` // string (simple) or array (verbose)
	IsError bool                `json:"is_error,omitempty"`
}

// claudeContentBlock represents a single content block in the verbose format.
type claudeContentBlock struct {
	Type    string `json:"type"`
	Text    string `json:"text,omitempty"`
	Name    string `json:"name,omitempty"`
	ID      string `json:"id,omitempty"`
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
			cost := ev.TotalCostUSD
			if cost == 0 {
				cost = ev.CostUSD
			}
			costMicros := int64(cost * 1_000_000)
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
		// Content may be a JSON string or a complex object; extract as string.
		var output string
		_ = json.Unmarshal(msg.Content, &output)
		return &agentv1.StreamEvent{
			Payload: &agentv1.StreamEvent_ToolResult{
				ToolResult: &agentv1.ToolResult{
					ToolName: msg.Name,
					Success:  !msg.IsError,
					Output:   output,
				},
			},
		}

	case "message":
		// Verbose format: content is an array of content blocks.
		return a.parseContentBlocks(msg.Content)

	default:
		return nil
	}
}

// parseContentBlocks extracts the first actionable event from a verbose content array.
func (a *ClaudeAdapter) parseContentBlocks(raw json.RawMessage) *agentv1.StreamEvent {
	var blocks []claudeContentBlock
	if err := json.Unmarshal(raw, &blocks); err != nil {
		return nil
	}

	// Collect all text blocks into a single text event.
	var texts []string
	for _, b := range blocks {
		switch b.Type {
		case "text":
			texts = append(texts, b.Text)
		case "tool_use":
			return &agentv1.StreamEvent{
				Payload: &agentv1.StreamEvent_ToolUse{
					ToolUse: &agentv1.ToolUse{ToolName: b.Name},
				},
			}
		case "tool_result":
			return &agentv1.StreamEvent{
				Payload: &agentv1.StreamEvent_ToolResult{
					ToolResult: &agentv1.ToolResult{
						ToolName: b.Name,
						Success:  !b.IsError,
						Output:   b.Content,
					},
				},
			}
		}
	}

	if len(texts) > 0 {
		return &agentv1.StreamEvent{
			Payload: &agentv1.StreamEvent_Text{Text: strings.Join(texts, "")},
		}
	}
	return nil
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
