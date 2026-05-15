package middleware

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/golang-jwt/jwt/v5"
)

// ContextKey is the type used for context keys set by this package.
type ContextKey string

// ClaimsKey is the context key under which JWT claims are stored.
const ClaimsKey ContextKey = "claims"

// GetUserIDFromContext extracts the authenticated user's ID from ctx.
// Returns an error if no claims are present or the user_id claim is
// missing or empty — callers should map this to a 401.
func GetUserIDFromContext(ctx context.Context) (string, error) {
	raw := ctx.Value(ClaimsKey)
	if raw == nil {
		return "", fmt.Errorf("no auth claims in context")
	}
	claims, ok := raw.(*jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("unexpected claims type in context")
	}
	userID, ok := (*claims)["user_id"].(string)
	if !ok || userID == "" {
		return "", fmt.Errorf("user_id not found in JWT claims")
	}
	return userID, nil
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// 1. Get token (cookie example)
		cookie, err := r.Cookie("auth_token")
		if err != nil {
			http.Error(w, "missing token", http.StatusUnauthorized)
			return
		}

		// 2. Validate token
		claims, err := ValidateJWT(cookie.Value)
		if err != nil {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), ClaimsKey, claims)

		// 3. Store claims in request context
		r = r.WithContext(ctx)

		// 4. Continue to handler
		next.ServeHTTP(w, r)
	})
}

func ValidateJWT(tokenString string) (*jwt.MapClaims, error) {
	claims := jwt.MapClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return &claims, nil
}
