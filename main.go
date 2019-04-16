package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"html/template"
	"net/http"
)

var homeTemplate *template.Template
var contactusTemplate *template.Template

func IndexHandler(w http.ResponseWriter, r *http.Request)  {
	w.Header().Set("Content-Type", "text/html")
	if err := homeTemplate.Execute(w, nil); err != nil {
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

	var err error
	homeTemplate, err = template.ParseFiles("views/home.gohtml","views/layouts/footer.gohtml")
	if err != nil {
		panic(err)
	}

	contactusTemplate, err = template.ParseFiles("views/contact.gohtml","views/layouts/footer.gohtml")
	if err != nil {
		panic(err)
	}

	r := mux.NewRouter()

	var h http.Handler = http.HandlerFunc(CustomNotFound)

	r.NotFoundHandler = h

	r.HandleFunc("/", IndexHandler)
	r.HandleFunc("/hello/{name}", GreetHandler)
	r.HandleFunc("/contactus", ContactUsHandler)

	http.ListenAndServe(":3000", r)
}

func ContactUsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	if err := contactusTemplate.Execute(w, nil); err != nil {
		panic(err)
	}
}
