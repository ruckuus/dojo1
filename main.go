package main

import (
	"fmt"
	"net/http"
)

func handlerFunc(w http.ResponseWriter, r *http.Request)  {
	w.Header().Set("Content-Type", "text/html")
	if r.URL.Path == "/" {
		fmt.Fprintf(w, "This is a homepage")
	} else if r.URL.Path == "/hello" {
		fmt.Fprintf(w, "Please contact: ruckuus@gmail.com")
	} else {
		w.WriteHeader(404)
		fmt.Fprintf(w, "Under construction")
	}
}

func main()  {
	mux := &http.ServeMux{}
	mux.HandleFunc("/", handlerFunc)
	http.ListenAndServe(":3000", mux)
}
