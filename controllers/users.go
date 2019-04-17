package controllers

import (
	"github.com/ruckuus/dojo1/views"
	"net/http"
)

type Users struct {
	NewView *views.View
}

type SignupForm struct {
	Name     string `schema:"name"`
	Email    string `schema:"email"`
	Password string `schema:"password"`
}

func NewUsers() *Users {
	return &Users{
		NewView: views.NewView("bootstrap", "users/new"),
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
		panic(err)
	}
}
