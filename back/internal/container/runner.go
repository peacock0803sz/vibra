package container

import (
	"context"
	"fmt"
	"io"

	"github.com/peacock0803sz/vibra/back/internal/adapter"
)

// Runtime abstracts the container runtime (Docker, Podman, etc.).
type Runtime interface {
	Ping(ctx context.Context) error
	Create(ctx context.Context, spec *adapter.ContainerSpec) (string, error)
	Start(ctx context.Context, id string) error
	// Attach connects to the container's stdout/stderr stream.
	// Must be called BEFORE Start to avoid missing early output.
	Attach(ctx context.Context, id string) (io.ReadCloser, error)
	Kill(ctx context.Context, id string) error
	Remove(ctx context.Context, id string) error
}

// Runner manages the lifecycle of agent containers.
type Runner struct {
	runtime Runtime
}

// NewRunner creates a Runner backed by the given container runtime.
func NewRunner(rt Runtime) *Runner {
	return &Runner{runtime: rt}
}

// Run starts a container from the spec and returns a reader for stdout.
// The container is automatically killed and removed when ctx is cancelled.
func (r *Runner) Run(ctx context.Context, spec *adapter.ContainerSpec) (io.ReadCloser, error) {
	id, err := r.runtime.Create(ctx, spec)
	if err != nil {
		return nil, fmt.Errorf("create container: %w", err)
	}

	// Attach before Start so we don't miss early output.
	logs, err := r.runtime.Attach(ctx, id)
	if err != nil {
		_ = r.runtime.Remove(ctx, id)
		return nil, fmt.Errorf("attach container: %w", err)
	}

	if err := r.runtime.Start(ctx, id); err != nil {
		logs.Close()
		_ = r.runtime.Remove(ctx, id)
		return nil, fmt.Errorf("start container: %w", err)
	}

	// Kill and remove container when context is cancelled.
	go func() {
		<-ctx.Done()
		_ = r.runtime.Kill(context.Background(), id)
		_ = r.runtime.Remove(context.Background(), id)
	}()

	return logs, nil
}
