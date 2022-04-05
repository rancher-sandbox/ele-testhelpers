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
			_, err := w.Write([]byte(content))
			if err != nil {
				fmt.Printf("Write failed: %v\n", err)
			}
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
