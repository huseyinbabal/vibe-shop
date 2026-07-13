package auth

import (
	"context"
	"net/http"
	"strings"

	"vibe-shop/internal/httpx"
)

// contextKey is unexported so only this package can set or read the user id in
// a request context.
type contextKey int

const userIDKey contextKey = 0

// UserIDFromContext returns the authenticated user id placed by RequireAuth,
// or ok=false when the request was not authenticated.
func UserIDFromContext(ctx context.Context) (uint, bool) {
	id, ok := ctx.Value(userIDKey).(uint)
	return id, ok
}

// RequireAuth wraps next so it runs only for requests carrying a valid
// "Authorization: Bearer <jwt>" header. The verified user id is stored in the
// request context; missing or invalid tokens get a 401.
func (tm *TokenManager) RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
		if !ok || token == "" {
			httpx.WriteError(w, http.StatusUnauthorized, "missing or malformed authorization header")
			return
		}

		userID, err := tm.Parse(token)
		if err != nil {
			httpx.WriteError(w, http.StatusUnauthorized, "invalid or expired token")
			return
		}

		ctx := context.WithValue(r.Context(), userIDKey, userID)
		next(w, r.WithContext(ctx))
	}
}
