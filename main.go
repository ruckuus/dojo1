package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/ruckuus/dojo1/views"
	"net/http"
)

var homeView *views.View
var contactView *views.View

func home(w http.ResponseWriter, r *http.Request)  {
	w.Header().Set("Content-Type", "text/html")

	data := struct {
		SiteName string
	}{"HomeButler"}
	if err := homeView.Render(w, data); err != nil {
		panic(err)
	}
}

func contact(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	if err := contactView.Render(w, nil); err != nil {
		panic(err)
	}
}

func CustomNotFound(w http.ResponseWriter, r *http.Request)  {
	fmt.Fprintf(w, "Could not find the page you're looking for ... ")
}

func main()  {

	homeView = views.NewView("bootstrap", "views/home.gohtml")
	contactView = views.NewView("bootstrap", "views/contact.gohtml")

	r := mux.NewRouter()

	var h http.Handler = http.HandlerFunc(CustomNotFound)

	r.NotFoundHandler = h

	r.HandleFunc("/", home)
	r.HandleFunc("/contact", contact)

	http.ListenAndServe(":3000", r)
}

