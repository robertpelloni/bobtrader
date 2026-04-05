package httpapi

import (
	"context"
	"errors"
	"fmt"
	"net/http"
)

type Runtime struct {
	server *http.Server
}

func NewRuntime(address string, handler http.Handler) *Runtime {
	return &Runtime{server: &http.Server{Addr: address, Handler: handler}}
}

func (r *Runtime) Start(ctx context.Context) error {
	go func() {
		<-ctx.Done()
		_ = r.server.Shutdown(context.Background())
	}()

	go func() {
		if err := r.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			fmt.Printf("http runtime error: %v\n", err)
		}
	}()
	return nil
}
