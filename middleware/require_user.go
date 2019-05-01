package middleware

import (
	"github.com/ruckuus/dojo1/context"
	"net/http"
)

type RequireUser struct{}

func (ru *RequireUser) ApplyFn(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := context.User(r.Context())
		if user == nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		next(w, r)
	})
}

func (ru *RequireUser) Apply(next http.Handler) http.HandlerFunc {
	return ru.ApplyFn(next.ServeHTTP)
}
