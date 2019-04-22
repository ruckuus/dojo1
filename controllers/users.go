package controllers

import (
	"fmt"
	"github.com/ruckuus/dojo1/models"
	"github.com/ruckuus/dojo1/rand"
	"github.com/ruckuus/dojo1/views"
	"net/http"
)

type Users struct {
	NewView   *views.View
	LoginView *views.View
	us        models.UserService
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

func NewUsers(us models.UserService) *Users {
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
	alert := views.Alert{
		Level:   "success",
		Message: "Successfully rendered dynamic alert",
	}

	data := views.Data{
		Alert: &alert,
		Yield: "Any data since it's an interface",
	}

	if err := u.NewView.Render(w, data); err != nil {
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

	err = u.signIn(w, &user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/cookietest", http.StatusFound)
}

func (u *Users) Login(w http.ResponseWriter, r *http.Request) {
	form := LoginForm{}
	err := parseForm(r, &form)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	foundUser, err := u.us.Authenticate(form.Email, form.Password)

	if err != nil {
		switch err {
		case models.ErrNotFound:
			http.Error(w, err.Error(), http.StatusNotFound)
		case models.ErrPasswordInvalid:
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}

	err = u.signIn(w, foundUser)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/cookietest", http.StatusFound)

}

func (u *Users) signIn(w http.ResponseWriter, user *models.User) error {
	if user.Remember == "" {
		token, err := rand.RememberToken()
		if err != nil {
			return err
		}
		user.Remember = token
		err = u.us.Update(user)
		if err != nil {
			return err
		}
	}
	cookie := http.Cookie{
		Name:     "remember_token",
		Value:    user.Remember,
		HttpOnly: true,
	}

	http.SetCookie(w, &cookie)
	return nil
}

func (u *Users) CookieTest(w http.ResponseWriter, r *http.Request) {
	cookie, _ := r.Cookie("remember_token")
	user, err := u.us.ByRemember(cookie.Value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
	}
	fmt.Fprintf(w, "Found user: %+v", user)
}
