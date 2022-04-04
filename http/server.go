package http

import (
	"context"
	"fmt"
	"net/http"
)

func Server(ctx context.Context, listenAddr string, content string) {
	srv := http.Server{
		Addr: listenAddr,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(content))
		}),
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			fmt.Printf("Server failed: %s\n", err)
		}
	}()
	go func() {
		<-ctx.Done()
		srv.Close()
	}()
}
