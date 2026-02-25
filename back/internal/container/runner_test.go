package container

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/peacock0803sz/vibra/back/internal/adapter"
)

// mockRuntime is a test implementation of the Runtime interface.
type mockRuntime struct {
	mu         sync.Mutex
	containers map[string]*mockContainer
	pingErr    error
	createErr  error
	startErr   error
}

type mockContainer struct {
	spec    *adapter.ContainerSpec
	started bool
	killed  bool
	removed bool
	logs    *bytes.Buffer
}

func newMockRuntime() *mockRuntime {
	return &mockRuntime{
		containers: make(map[string]*mockContainer),
	}
}

func (m *mockRuntime) Ping(ctx context.Context) error {
	return m.pingErr
}

func (m *mockRuntime) Create(ctx context.Context, spec *adapter.ContainerSpec) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.createErr != nil {
		return "", m.createErr
	}
	id := fmt.Sprintf("container-%d", len(m.containers)+1)
	m.containers[id] = &mockContainer{
		spec: spec,
		logs: bytes.NewBufferString("line1\nline2\nline3\n"),
	}
	return id, nil
}

func (m *mockRuntime) Start(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.startErr != nil {
		return m.startErr
	}
	c, ok := m.containers[id]
	if !ok {
		return fmt.Errorf("container not found: %s", id)
	}
	c.started = true
	return nil
}

func (m *mockRuntime) Attach(ctx context.Context, id string) (io.ReadCloser, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	c, ok := m.containers[id]
	if !ok {
		return nil, fmt.Errorf("container not found: %s", id)
	}
	return io.NopCloser(c.logs), nil
}

func (m *mockRuntime) Kill(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if c, ok := m.containers[id]; ok {
		c.killed = true
	}
	return nil
}

func (m *mockRuntime) Remove(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if c, ok := m.containers[id]; ok {
		c.removed = true
	}
	return nil
}

// snapshot returns a copy of container state under the lock.
func (m *mockRuntime) snapshot() map[string]mockContainer {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make(map[string]mockContainer, len(m.containers))
	for id, c := range m.containers {
		result[id] = *c
	}
	return result
}

func TestRunner_Run(t *testing.T) {
	rt := newMockRuntime()
	runner := NewRunner(rt)

	spec := &adapter.ContainerSpec{
		Image:   "test-image:latest",
		Command: []string{"echo", "hello"},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	reader, err := runner.Run(ctx, spec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("read error: %v", err)
	}

	if string(data) != "line1\nline2\nline3\n" {
		t.Errorf("got %q, want %q", string(data), "line1\\nline2\\nline3\\n")
	}

	snap := rt.snapshot()
	if len(snap) != 1 {
		t.Fatalf("expected 1 container, got %d", len(snap))
	}
	for _, c := range snap {
		if !c.started {
			t.Error("container not started")
		}
	}
}

func TestRunner_KillOnCancel(t *testing.T) {
	rt := newMockRuntime()
	runner := NewRunner(rt)

	spec := &adapter.ContainerSpec{
		Image:   "test-image:latest",
		Command: []string{"sleep", "infinity"},
	}

	ctx, cancel := context.WithCancel(context.Background())
	_, err := runner.Run(ctx, spec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cancel()

	// Poll until the cleanup goroutine completes (max 1s).
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		snap := rt.snapshot()
		allDone := true
		for _, c := range snap {
			if !c.killed || !c.removed {
				allDone = false
				break
			}
		}
		if allDone {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}

	snap := rt.snapshot()
	for _, c := range snap {
		if !c.killed {
			t.Error("container not killed after context cancel")
		}
		if !c.removed {
			t.Error("container not removed after context cancel")
		}
	}
}

func TestRunner_CreateError(t *testing.T) {
	rt := newMockRuntime()
	rt.createErr = fmt.Errorf("image not found")
	runner := NewRunner(rt)

	spec := &adapter.ContainerSpec{Image: "missing:latest"}
	_, err := runner.Run(context.Background(), spec)
	if err == nil {
		t.Fatal("expected error for create failure")
	}
}

func TestRunner_StartError(t *testing.T) {
	rt := newMockRuntime()
	rt.startErr = fmt.Errorf("start failed")
	runner := NewRunner(rt)

	spec := &adapter.ContainerSpec{Image: "test:latest"}
	_, err := runner.Run(context.Background(), spec)
	if err == nil {
		t.Fatal("expected error for start failure")
	}

	// Verify container was removed after start failure.
	snap := rt.snapshot()
	for _, c := range snap {
		if !c.removed {
			t.Error("container should be removed after start failure")
		}
	}
}
