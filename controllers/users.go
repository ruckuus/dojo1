package controllers

import (
	"fmt"
	"github.com/ruckuus/dojo1/context"
	"github.com/ruckuus/dojo1/email"
	"github.com/ruckuus/dojo1/models"
	"github.com/ruckuus/dojo1/rand"
	"github.com/ruckuus/dojo1/views"
	"net/http"
	"time"
)

type Users struct {
	NewView      *views.View
	LoginView    *views.View
	ForgotPwView *views.View
	ResetPwView  *views.View
	us           models.UserService
	emailer      *email.Client
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

type ResetPwForm struct {
	Email    string `schema:"email"`
	Token    string `schema:"token"`
	Password string `schema:"password"`
}

func NewUsers(us models.UserService, emailer *email.Client) *Users {
	return &Users{
		NewView:      views.NewView("bootstrap", "users/new"),
		LoginView:    views.NewView("bootstrap", "users/login"),
		ForgotPwView: views.NewView("bootstrap", "users/forgot_pw"),
		ResetPwView:  views.NewView("bootstrap", "users/reset_pw"),
		us:           us,
		emailer:      emailer,
	}
}

// New is used to render the signup form
//
// GET /signup
func (u *Users) New(w http.ResponseWriter, r *http.Request) {
	var form SignupForm
	parseURLParams(r, &form)
	u.NewView.Render(w, r, form)

}

// New is used to create a new user
func (u *Users) Create(w http.ResponseWriter, r *http.Request) {
	var vd views.Data

	p := SignupForm{}

	// copy form input back to signupForm
	vd.Yield = &p

	err := parseForm(r, &p)

	if err != nil {
		vd.SetAlert(err)
		u.NewView.Render(w, r, vd)
		return
	}

	user := models.User{
		Name:     p.Name,
		Email:    p.Email,
		Password: p.Password,
	}

	if err = u.us.Create(&user); err != nil {
		vd.SetAlert(err)
		u.NewView.Render(w, r, vd)
		return
	}

	u.emailer.Welcome(user.Name, user.Email)
	err = u.signIn(w, &user)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	http.Redirect(w, r, "/galleries", http.StatusFound)
}

func (u *Users) Login(w http.ResponseWriter, r *http.Request) {
	var vd views.Data
	form := LoginForm{}
	err := parseForm(r, &form)

	if err != nil {
		vd.SetAlert(err)
		u.NewView.Render(w, r, vd)
		return
	}

	foundUser, err := u.us.Authenticate(form.Email, form.Password)

	if err != nil {
		switch err {
		case models.ErrNotFound:
			vd.Alert = &views.Alert{
				Level:   views.AlertLvlError,
				Message: "No such user with that email address.",
			}
		case models.ErrPasswordInvalid:
			vd.Alert = &views.Alert{
				Level:   views.AlertLvlError,
				Message: "Password is incorrect.",
			}
		default:
			vd.SetAlert(err)
		}
		u.LoginView.Render(w, r, vd)
		return
	}

	err = u.signIn(w, foundUser)
	if err != nil {
		vd.SetAlert(err)
		u.LoginView.Render(w, r, vd)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)

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

func (u *Users) Logout(w http.ResponseWriter, r *http.Request) {
	cookie := http.Cookie{
		Name:     "remember_token",
		Value:    "",
		Expires:  time.Now(),
		HttpOnly: true,
	}

	http.SetCookie(w, &cookie)

	user := context.User(r.Context())

	token, _ := rand.RememberToken()
	user.Remember = token

	u.us.Update(user)

	http.Redirect(w, r, "/", http.StatusFound)
}

func (u *Users) CookieTest(w http.ResponseWriter, r *http.Request) {
	cookie, _ := r.Cookie("remember_token")
	user, err := u.us.ByRemember(cookie.Value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
	}
	fmt.Fprintf(w, "Found user: %+v", user)
}

// POST /forgot
func (u *Users) InitiateReset(w http.ResponseWriter, r *http.Request) {
	var vd views.Data
	var form ResetPwForm
	vd.Yield = &form
	if err := parseForm(r, &form); err != nil {
		vd.SetAlert(err)
		u.ForgotPwView.Render(w, r, vd)
		return
	}

	token, err := u.us.InitiateReset(form.Email)
	fmt.Println("Token for: ", form.Email, " : ", token)

	if err != nil {
		vd.SetAlert(err)
		u.ForgotPwView.Render(w, r, vd)
		return
	}

	views.RedirectAlert(w, r, "/reset", http.StatusFound, views.Alert{
		Level:   views.AlertLvlSuccess,
		Message: "If we find the provided email, a confirmation link will be sent to that email address.",
	})
}

// GET /reset
func (u *Users) ResetPw(w http.ResponseWriter, r *http.Request) {
	var vd views.Data
	var form ResetPwForm
	vd.Yield = &form

	if err := parseForm(r, &form); err != nil {
		vd.SetAlert(err)
	}
	u.ResetPwView.Render(w, r, vd)
}

// POST /reset
func (u *Users) CompleteReset(w http.ResponseWriter, r *http.Request) {
	var vd views.Data
	var form ResetPwForm
	vd.Yield = &form
	if err := parseForm(r, &form); err != nil {
		vd.SetAlert(err)
		u.ResetPwView.Render(w, r, vd)
		return
	}

	user, err := u.us.CompleteReset(form.Token, form.Password)
	if err != nil {
		vd.SetAlert(err)
		u.ResetPwView.Render(w, r, vd)
		return
	}

	u.signIn(w, user)
	views.RedirectAlert(w, r, "/galleries", http.StatusFound, views.Alert{
		Level:   views.AlertLvlSuccess,
		Message: "Your password has been reset and you have been logged in.",
	})
}
