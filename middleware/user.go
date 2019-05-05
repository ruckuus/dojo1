package middleware

import (
	"github.com/ruckuus/dojo1/context"
	"github.com/ruckuus/dojo1/models"
	"net/http"
	"strings"
)

type User struct {
	models.UserService
}

func (u *User) ApplyFn(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		// If the user is requesting a static asset or image
		// we will not need to lookup the current user so we skip
		// doing that.
		if strings.HasPrefix(path, "/assets/") ||
			strings.HasPrefix(path, "/images/") {
			next(w, r)
			return
		}

		cookie, err := r.Cookie("remember_token")
		if err != nil {
			next(w, r)
			return
		}

		user, err := u.ByRemember(cookie.Value)

		if err != nil {
			next(w, r)
			return
		}

		ctx := r.Context()
		ctx = context.WithUser(ctx, user)
		r = r.WithContext(ctx)

		next(w, r)
	})
}

func (u *User) Apply(next http.Handler) http.HandlerFunc {
	return u.ApplyFn(next.ServeHTTP)
}
