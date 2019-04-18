package controllers

import (
	"fmt"
	"github.com/ruckuus/dojo1/models"
	"github.com/ruckuus/dojo1/views"
	"net/http"
)

type Users struct {
	NewView   *views.View
	LoginView *views.View
	us        *models.UserService
}

type SignupForm struct {
	Name     string `schema:"name"`
	Email    string `schema:"email"`
	Password string `schema:"password"`
}

type LoginForm struct {
	Email    string `schema:"email"`
	Password string `schema:"password"`
}

func NewUsers(us *models.UserService) *Users {
	return &Users{
		NewView:   views.NewView("bootstrap", "users/new"),
		LoginView: views.NewView("bootstrap", "users/login"),
		us:        us,
	}
}

// New is used to render the signup form
//
// GET /signup
func (u *Users) New(w http.ResponseWriter, r *http.Request) {
	if err := u.NewView.Render(w, nil); err != nil {
		panic(err)
	}
}

// New is used to create a new user
func (u *Users) Create(w http.ResponseWriter, r *http.Request) {
	p := SignupForm{}
	err := parseForm(r, &p)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	user := models.User{
		Name:     p.Name,
		Email:    p.Email,
		Password: p.Password,
	}

	if err = u.us.Create(&user); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (u *Users) Login(w http.ResponseWriter, r *http.Request) {
	form := LoginForm{}
	err := parseForm(r, &form)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	foundUser, err := u.us.Authenticate(form.Email, form.Password)

	switch err {
	case nil:
		fmt.Fprintf(w, "Found user: +%v", foundUser)
	case models.ErrNotFound:
		http.Error(w, err.Error(), http.StatusNotFound)
	case models.ErrInvalidPassword:
		http.Error(w, err.Error(), http.StatusBadRequest)
	default:
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}
