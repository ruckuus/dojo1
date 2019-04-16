package main

import (
	"github.com/gorilla/mux"
	"github.com/ruckuus/dojo1/views"
	"net/http"
)

var (
	homeView *views.View
	contactView *views.View
	errorView *views.View
	faqView *views.View
	signupView *views.View
)

func home(w http.ResponseWriter, r *http.Request)  {
	w.Header().Set("Content-Type", "text/html")

	data := struct {
		SiteName string
	}{"HomeButler"}
	must(homeView.Render(w, data))
}

func contact(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	must(contactView.Render(w, nil))
}

func CustomNotFound(w http.ResponseWriter, r *http.Request)  {
	w.Header().Set("Content-Type", "text/html")
	errorData := struct {
		ErrorCode int
		ErrorMessage string
	} {
		404, "Could not find what you're looking for ...",
	}
	must(errorView.Render(w, errorData))
}

func faq(w http.ResponseWriter, r *http.Request)  {
	w.Header().Set("Content-Type", "text/html")
	must(faqView.Render(w, nil))
}

func main()  {

	homeView = views.NewView("bootstrap", "views/home.gohtml")
	contactView = views.NewView("bootstrap", "views/contact.gohtml")
	errorView = views.NewView("bootstrap", "views/error.gohtml")
	faqView = views.NewView("bootstrap", "views/faq.gohtml")
	signupView = views.NewView("bootstrap", "views/signup.gohtml")

	r := mux.NewRouter()

	var h http.Handler = http.HandlerFunc(CustomNotFound)

	r.NotFoundHandler = h

	r.HandleFunc("/", home)
	r.HandleFunc("/contact", contact)
	r.HandleFunc("/faq", faq)
	r.HandleFunc("/signup", signup)

	http.ListenAndServe(":3000", r)
}

func signup(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	must(signupView.Render(w, nil))
}

func must(err error)  {
	if err != nil {
		panic(err)
	}
}

