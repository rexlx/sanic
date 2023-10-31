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
		http.ServeFile(w, r, path)
	}
	return &f, nil
}
