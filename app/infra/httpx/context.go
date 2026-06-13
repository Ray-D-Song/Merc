package httpx

import (
	"context"
	"errors"
	"net/http"
)

const ContextUserIDKey = "user_id"

func GetUserID(r *http.Request) (uint, error) {
	userID, ok := r.Context().Value(ContextUserIDKey).(uint)
	if !ok {
		return 0, errors.New("user not authenticated")
	}
	return userID, nil
}

func SetUserID(ctx context.Context, userID uint) context.Context {
	return context.WithValue(ctx, ContextUserIDKey, userID)
}
