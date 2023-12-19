package context

import (
	"context"

	"github.com/taherk/galleryapp/models"
)

// define types for context keys to avoid conflicts with other
// packages using the same keys for setting context values.
// context.WithValue retrives value on the basis of the type and
// value of the key.
type key string

const (
	userKey key = "user"
)

func WithUser(ctx context.Context, user *models.User) context.Context {
	return context.WithValue(ctx, userKey, user)
}

func User(ctx context.Context) *models.User {
	val := ctx.Value(userKey)
	user, ok := val.(*models.User)
	if !ok {
		// the most likely case is that nothing was ever stored in the contex,
		// so it dosen't have a type of *models.User. It is also possible that
		// other code in this package wrote an invalid value using the user key.
		return nil
	}

	return user
}
