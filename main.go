package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

func IndexHandler(w http.ResponseWriter, r *http.Request)  {
	fmt.Fprintf(w, "This is homepage")
}

func GreetHandler(w http.ResponseWriter, r *http.Request)  {
	vars := mux.Vars(r)
	name := vars["name"]
	fmt.Fprintf(w, "Hello, %v", name)
}

func CustomNotFound(w http.ResponseWriter, r *http.Request)  {
	fmt.Fprintf(w, "Could not find the page you're looking for ... ")
}

func main()  {
	r := mux.NewRouter()

	var h http.Handler = http.HandlerFunc(CustomNotFound)

	r.NotFoundHandler = h

	r.HandleFunc("/", IndexHandler)
	r.HandleFunc("/hello/{name}", GreetHandler)

	http.ListenAndServe(":3000", r)
}
