package httpapi

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"
)

type Runtime struct {
	server   *http.Server
	listener net.Listener
	mu       sync.Mutex
}

func NewRuntime(address string, handler http.Handler) *Runtime {
	return &Runtime{server: &http.Server{Addr: address, Handler: handler}}
}

func (r *Runtime) Start(ctx context.Context) error {
	r.mu.Lock()
	if r.listener != nil {
		r.mu.Unlock()
		return nil
	}
	ln, err := net.Listen("tcp", r.server.Addr)
	if err != nil {
		r.mu.Unlock()
		return fmt.Errorf("listen: %w", err)
	}
	r.listener = ln
	r.mu.Unlock()

	go func() {
		<-ctx.Done()
		_ = r.Shutdown(context.Background())
	}()
	go func() {
		if err := r.server.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
			fmt.Printf("http runtime error: %v\n", err)
		}
	}()
	return nil
}

func (r *Runtime) Shutdown(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.listener == nil {
		return nil
	}
	err := r.server.Shutdown(ctx)
	r.listener = nil
	return err
}

func (r *Runtime) Address() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.listener != nil {
		return r.listener.Addr().String()
	}
	return r.server.Addr
}
