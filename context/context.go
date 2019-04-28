package context

import (
	"context"
	"github.com/ruckuus/dojo1/models"
)

const userKey = "user"

func WithUser(ctx context.Context, user *models.User) context.Context {
	return context.WithValue(ctx, userKey, user)
}

func User(ctx context.Context) *models.User {
	if tmp := ctx.Value(userKey); tmp != nil {
		if user, ok := tmp.(*models.User); ok {
			return user
		}
	}
	return nil
}
