package main

import (
	"flag"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"github.com/ruckuus/dojo1/controllers"
	"github.com/ruckuus/dojo1/email"
	"github.com/ruckuus/dojo1/middleware"
	"github.com/ruckuus/dojo1/models"
	"github.com/ruckuus/dojo1/rand"
	"log"
	"net/http"
)

func main() {

	runProd := flag.Bool("prod", false, "Ensure app running with production config (.config)")

	config := LoadConfig(*runProd)

	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(config.AWSConfig.Region),
	}))

	services, err := models.NewServices(
		models.WithGorm(config.Database.Dialect(), config.Database.ConnectionInfo()),
		models.WithLogMode(!config.IsProd()),
		models.WithUser(config.Pepper, config.HMACKey),
		models.WithAWSSession(sess),
		models.WithS3Bucket(config.AWSConfig.Bucket),
		models.WithS3Store(),
		models.WithGallery(),
		models.WithImageCDNDomain(config.ImageCDNDomain),
		models.WithImage(),
		models.WithProperty(),
	)

	mailConfig := config.Mailgun
	emailer := email.NewClient(
		email.WithSender("Dwi from Tataruma", "hello@tataruma.com"),
		email.WithMailgun(mailConfig.Domain, mailConfig.APIKey),
	)

	if err != nil {
		panic(err)
	}

	defer services.Close()
	services.AutoMigrate()

	r := mux.NewRouter()

	// Middlewares

	// User middleware
	userMw := middleware.User{
		UserService: services.User,
	}
	requireUserMw := middleware.RequireUser{}

	// CSRF Middleware
	csrfKey, err := rand.Bytes(32)
	if err != nil {
		panic(err)
	}

	csrfMw := csrf.Protect(csrfKey, csrf.Secure(config.IsProd()))

	// Controllers
	userC := controllers.NewUsers(services.User, emailer)
	staticC := controllers.NewStatic()
	galleriesC := controllers.NewGalleries(services.Gallery, r, services.Image)
	propertiesC := controllers.NewProperties(services.Property, r)

	newGallery := requireUserMw.Apply(galleriesC.NewView)
	createGallery := requireUserMw.ApplyFn(galleriesC.Create)

	// Static router

	r.Handle("/", staticC.Home).Methods("GET")
	r.Handle("/contact", staticC.Contact).Methods("GET")
	r.Handle("/faq", staticC.FAQ).Methods("GET")

	// User router

	r.HandleFunc("/signup", userC.New).Methods("GET")
	r.HandleFunc("/signup", userC.Create).Methods("POST")
	r.Handle("/login", userC.LoginView).Methods("GET")
	r.HandleFunc("/login", userC.Login).Methods("POST")
	r.HandleFunc("/logout", requireUserMw.ApplyFn(userC.Logout)).Methods("POST")
	r.HandleFunc("/cookietest", userC.CookieTest).Methods("GET")
	r.Handle("/forgot", userC.ForgotPwView).Methods("GET")
	r.HandleFunc("/forgot", userC.InitiateReset).Methods("POST")
	r.HandleFunc("/reset", userC.ResetPw).Methods("GET")
	r.HandleFunc("/reset", userC.CompleteReset).Methods("POST")
	r.HandleFunc("/profile", userC.Profile).Methods("GET")

	// Gallery router
	r.Handle("/galleries/new", newGallery).Methods("GET")
	r.HandleFunc("/galleries", createGallery).Methods("POST")
	r.HandleFunc("/galleries/{id:[0-9]+}", galleriesC.Show).
		Methods("GET").
		Name(controllers.ShowGallery)
	r.HandleFunc("/galleries", requireUserMw.ApplyFn(galleriesC.Index)).
		Methods("GET").
		Name(controllers.IndexGalleries)
	r.HandleFunc("/galleries/{id:[0-9]+}/edit", requireUserMw.ApplyFn(galleriesC.Edit)).
		Methods("GET").
		Name(controllers.EditGallery)
	r.HandleFunc("/galleries/{id:[0-9]+}/update", requireUserMw.ApplyFn(galleriesC.Update)).
		Methods("POST").
		Name(controllers.UpdateGallery)
	r.HandleFunc("/galleries/{id:[0-9]+}/delete", requireUserMw.ApplyFn(galleriesC.Delete)).
		Methods("POST").
		Name(controllers.DeleteGallery)

	// Image Upload
	r.HandleFunc("/galleries/{id:[0-9]+}/images", requireUserMw.ApplyFn(galleriesC.ImageUpload)).
		Methods("POST")

	// Image Delete
	r.HandleFunc("/galleries/{id:[0-9]+}/images/{filename}/delete", requireUserMw.ApplyFn(galleriesC.ImageDelete)).
		Methods("POST")

	// Image routes
	imageHandler := http.FileServer(http.Dir("./images/"))
	r.PathPrefix("/images/").Handler(http.StripPrefix("/images/", imageHandler))

	// Static assets
	assetsHandler := http.FileServer(http.Dir("./public"))
	r.PathPrefix("/assets/").Handler(assetsHandler)

	// userMw.Apply(r) lets User Middleware to execute before the routes

	// Properties router
	r.HandleFunc("/properties/new", requireUserMw.ApplyFn(propertiesC.New)).
		Methods("GET")
	r.HandleFunc("/properties", requireUserMw.ApplyFn(propertiesC.Create)).
		Methods("POST")
	r.HandleFunc("/properties", requireUserMw.ApplyFn(propertiesC.Index)).
		Methods("GET").Name(controllers.IndexProperties)
	r.HandleFunc("/properties/{id:[0-9]+}", requireUserMw.ApplyFn(propertiesC.Show)).
		Methods("GET")
	r.HandleFunc("/properties/{id:[0-9]+}/edit", requireUserMw.ApplyFn(propertiesC.Edit)).
		Methods("GET")

	r.HandleFunc("/properties/{id:[0-9]+}/update", requireUserMw.ApplyFn(propertiesC.Update)).
		Methods("POST")
	r.HandleFunc("/properties/{id:[0-9]+}/delete", requireUserMw.ApplyFn(propertiesC.Delete)).
		Methods("POST")

	// End of properties router
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", config.Port), csrfMw(userMw.Apply(r))))
}
