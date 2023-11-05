package main

import (
	"fmt"
	"net/http"
	"os"
)

func AdditionalHandler(w http.ResponseWriter, r *http.Request) {
	out := `<p><em>wow</em>, look at you</p>`
	fmt.Fprintf(w, out)
}

func FaviconHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "mycon")
}

func UIServer(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "ui server")

}

func NewUIServer(path string) (*http.HandlerFunc, error) {
	var f http.HandlerFunc

	if path == "" {
		return &f, fmt.Errorf("path cannot be empty")
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return &f, fmt.Errorf("path %v does not exist", path)
	}

	f = func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/javascript")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		http.ServeFile(w, r, path)
		fmt.Println("served file", w.Header().Get("Content-Type"))
	}
	return &f, nil
}

func CorsHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("cors handler called")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		next.ServeHTTP(w, r)
	})
}
