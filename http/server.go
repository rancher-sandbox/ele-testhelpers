/*
Copyright © 2022 - 2025 SUSE LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package http

import (
	"context"
	"fmt"
	"net/http"
)

//nolint:all
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
