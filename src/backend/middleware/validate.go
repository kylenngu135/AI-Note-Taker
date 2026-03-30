package middleware

import (
	"net/http"
	"github.com/golang-jwt/jwt/v5"
	"fmt"
	"os"
	"context"
)

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
	
 		ctx := context.WithValue(r.Context(), "claims", claims)

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
