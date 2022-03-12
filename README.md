## Gemini Server
A simple Go library to build [Gemini](https://gemini.circumlunar.space/) web servers.
  

### Example
The basic structure is similar to `net/http`.
```go
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
```
Gemini clients allow self-signed certificates.  
You can use any client to view the contents. My personal favourite is [Amfora](https://github.com/makeworld-the-better-one/amfora).