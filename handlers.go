package main

import (
	"fmt"
	"net/http"
)

func AdditionalHandler(w http.ResponseWriter, r *http.Request) {
	out := `<p><em>wow</em>, look at you</p>`
	fmt.Fprintf(w, out)
}

func FaviconHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "mycon")
}
