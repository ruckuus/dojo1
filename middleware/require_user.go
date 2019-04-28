package middleware

import (
	"fmt"
	"github.com/ruckuus/dojo1/context"
	"github.com/ruckuus/dojo1/models"
	"net/http"
)

type RequireUser struct {
	models.UserService
}

func (ru *RequireUser) ApplyFn(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("remember_token")
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		user, err := ru.UserService.ByRemember(cookie.Value)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		fmt.Println("Found user: ", user)

		// Get context from request
		ctx := r.Context()

		// create a new context from existing one with our user
		ctx = context.WithUser(ctx, user)

		// create a new request from context with user attached,
		// assign back to r
		r = r.WithContext(ctx)

		next(w, r)
	})
}

func (ru *RequireUser) Apply(next http.Handler) http.HandlerFunc {
	return ru.ApplyFn(next.ServeHTTP)
}
