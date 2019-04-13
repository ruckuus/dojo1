package main

import (
	"fmt"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func Index(w http.ResponseWriter, r *http.Request, ps httprouter.Params)  {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, "This is homepage")
}

func Hello(w http.ResponseWriter, r *http.Request, ps httprouter.Params)  {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, "hello, %s!\n", ps.ByName("name"))
}

func main()  {
	router := httprouter.New()
	router.GET("/", Index)
	router.GET("/hello/:name", Hello)
	http.ListenAndServe(":3000", router)
}
