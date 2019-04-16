package main

import (
	"github.com/gorilla/mux"
	"github.com/ruckuus/dojo1/controllers"
	"net/http"
)

func main()  {

	userC := controllers.NewUsers()
	staticC := controllers.NewStatic()

	r := mux.NewRouter()

	r.Handle("/", staticC.Home).Methods("GET")
	r.Handle("/contact", staticC.Contact).Methods("GET")
	r.Handle("/faq", staticC.FAQ).Methods("GET")
	r.HandleFunc("/signup", userC.New).Methods("GET")
	r.HandleFunc("/signup", userC.Create).Methods("POST")

	http.ListenAndServe(":3000", r)
}

