package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/ruckuus/dojo1/views"
	"net/http"
)

var homeView *views.View
var contactView *views.View

func IndexHandler(w http.ResponseWriter, r *http.Request)  {
	w.Header().Set("Content-Type", "text/html")

	data := struct {
		SiteName string
	}{"HomeButler"}
	if err := homeView.Template.Execute(w, data); err != nil {
		panic(err)
	}
}

func ContactUsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	if err := contactView.Template.Execute(w, nil); err != nil {
		panic(err)
	}
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

	homeView = views.NewView("views/home.gohtml")
	contactView = views.NewView("views/contact.gohtml")

	r := mux.NewRouter()

	var h http.Handler = http.HandlerFunc(CustomNotFound)

	r.NotFoundHandler = h

	r.HandleFunc("/", IndexHandler)
	r.HandleFunc("/hello/{name}", GreetHandler)
	r.HandleFunc("/contactus", ContactUsHandler)

	http.ListenAndServe(":3000", r)
}

