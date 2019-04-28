package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/ruckuus/dojo1/controllers"
	"github.com/ruckuus/dojo1/middleware"
	"github.com/ruckuus/dojo1/models"
	"net/http"
)

const (
	host     = "localhost"
	port     = 54320
	user     = "lenslocked_user"
	password = "lenslocked_password"
	db_name  = "lenslocked_db"
)

func main() {

	// Prepare DB connection string
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, db_name)

	services, err := models.NewServices(psqlInfo)

	if err != nil {
		panic(err)
	}

	defer services.Close()
	services.AutoMigrate()

	// Middleware

	requireUserMw := middleware.RequireUser{
		UserService: services.User,
	}

	userC := controllers.NewUsers(services.User)
	staticC := controllers.NewStatic()
	galleriesC := controllers.NewGalleries(services.Gallery)

	newGallery := requireUserMw.Apply(galleriesC.NewView)
	createGallery := requireUserMw.ApplyFn(galleriesC.Create)

	r := mux.NewRouter()

	// Static router

	r.Handle("/", staticC.Home).Methods("GET")
	r.Handle("/contact", staticC.Contact).Methods("GET")
	r.Handle("/faq", staticC.FAQ).Methods("GET")

	// User router

	r.HandleFunc("/signup", userC.New).Methods("GET")
	r.HandleFunc("/signup", userC.Create).Methods("POST")
	r.Handle("/login", userC.LoginView).Methods("GET")
	r.HandleFunc("/login", userC.Login).Methods("POST")
	r.HandleFunc("/cookietest", userC.CookieTest).Methods("GET")

	// Gallery router
	r.Handle("/galleries/new", newGallery).Methods("GET")
	r.HandleFunc("/galleries", createGallery).Methods("POST")
	r.HandleFunc("/galleries/{id:[0-9]+}", galleriesC.Show).Methods("GET")

	http.ListenAndServe(":3000", r)
}
