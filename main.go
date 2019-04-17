package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/ruckuus/dojo1/controllers"
	"github.com/ruckuus/dojo1/models"
	"net/http"
)

const (
	host     = "localhost"
	port     = 32769
	user     = "lenslocked_db_user"
	password = "db_Password123!"
	db_name  = "lenslocked_db"
)

func main() {

	// Prepare DB connection string
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, db_name)

	us, err := models.NewUserService(psqlInfo)
	if err != nil {
		panic(err)
	}
	defer us.Close()
	us.AutoMigrate()

	userC := controllers.NewUsers(us)
	staticC := controllers.NewStatic()
	galleriesC := controllers.NewGalleries()

	r := mux.NewRouter()

	r.Handle("/", staticC.Home).Methods("GET")
	r.Handle("/contact", staticC.Contact).Methods("GET")
	r.Handle("/faq", staticC.FAQ).Methods("GET")
	r.HandleFunc("/signup", userC.New).Methods("GET")
	r.HandleFunc("/signup", userC.Create).Methods("POST")
	r.HandleFunc("/galleries/new", galleriesC.New).Methods("GET")

	http.ListenAndServe(":3000", r)
}
