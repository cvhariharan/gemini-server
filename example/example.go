package main

import (
	"log"

	"github.com/cvhariharan/gemini-server"
)

func main() {
	gemini.HandleFunc("/", func(w *gemini.Response, r *gemini.Request) {
		w.SetStatus(gemini.StatusSuccess, "text/gemini")
		w.Write([]byte("# Test Response"))
	})

	log.Fatal(gemini.ListenAndServeTLS(":1965", "localhost.crt", "localhost.key"))
}
