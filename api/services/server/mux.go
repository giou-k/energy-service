package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

// WebAPI defines a dummy multiplexer.
func WebAPI() *http.ServeMux {
	mux := http.NewServeMux()

	h := func(w http.ResponseWriter, req *http.Request) {
		status := struct {
			Status string
		}{
			Status: "OK",
		}

		err := json.NewEncoder(w).Encode(status)
		if err != nil {
			fmt.Printf("error in encode: %w", err)
			os.Exit(1)
		}
	}

	h2 := func(w http.ResponseWriter, req *http.Request) {
		err := json.NewEncoder(w).Encode("THIS is a message")
		if err != nil {
			fmt.Printf("error in encode: %w", err)
			os.Exit(1)
		}
	}

	mux.HandleFunc("/", h2)
	mux.HandleFunc("/test", h)

	return mux
}
