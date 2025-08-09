package handlers

import (
	"net/http"
	"os"
	"strings"
	"context"

	"github.com/golang-jwt/jwt/v5"
)

type ctxKey string
const userKey ctxKey = "authUserID"

func UserIDFromCtx(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(userKey).(string)
	return v, ok
}

func AuthMiddleware(next http.Handler) http.Handler {
	secret := os.Getenv("JWT_SECRET")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		tokenStr := strings.TrimPrefix(auth, "Bearer ")
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
			if t.Method != jwt.SigningMethodHS256 {
				return nil, jwt.ErrTokenUnverifiable
			}
			return []byte(secret), nil
		})
		if err != nil || !token.Valid {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		claims, _ := token.Claims.(jwt.MapClaims)
		uid, _ := claims["sub"].(string) // or "uid" depending on your issuer; adjust to your MakeJWT
		if uid == "" {
			// fallback: try "uid"
			if v, ok := claims["uid"].(string); ok {
				uid = v
			}
		}
		if uid == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), userKey, uid)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
