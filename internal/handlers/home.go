package handlers

import "net/http"

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte("Hello World with air Reload!"))
}
