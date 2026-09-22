package main

import (
	"fmt"
	"net/http"
	"os"
	"time"
)

func main() {
	port := os.Args[1]

	delay := 10 * time.Millisecond
	if port == "8000" {
		delay = 100 * time.Millisecond
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(delay)
		fmt.Fprintf(w, "hello from backend %s\n", port)
	})

	fmt.Println("backend started on", port, "with delay", delay)
	http.ListenAndServe(":"+port, nil)
}
