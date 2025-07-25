package handlers

import (
	"context"
	"net/http"
	"strings"

	"cctv-api/internal/responses"
	"cctv-api/internal/utils"
)

type contextKey string

const (
	userClaimsKey contextKey = "userClaims"
)

func JWTMiddleware(jwtUtil *utils.JWTUtil) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip middleware untuk endpoint tertentu
			if r.URL.Path == "/api/auth/login" ||
				r.URL.Path == "/api/auth/register" ||
				r.URL.Path == "/api/health" {
				next.ServeHTTP(w, r)
				return
			}

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				responses.SendErrorResponse(w, http.StatusUnauthorized, "Authorization header is required")
				return
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString == authHeader {
				responses.SendErrorResponse(w, http.StatusUnauthorized, "Bearer token not found")
				return
			}

			claims, err := jwtUtil.ValidateToken(tokenString)
			if err != nil {
				responses.SendErrorResponse(w, http.StatusUnauthorized, "Invalid token: "+err.Error())
				return
			}

			// Verify token matches the one in database
			var dbToken string
			err = jwtUtil.DB.QueryRow("SELECT session_token FROM users WHERE id = $1", claims.UserID).Scan(&dbToken)
			if err != nil || dbToken != tokenString {
				responses.SendErrorResponse(w, http.StatusUnauthorized, "Token mismatch - possibly logged in from another device")
				return
			}

			ctx := context.WithValue(r.Context(), userClaimsKey, claims)
			ctx = context.WithValue(ctx, "userID", claims.UserID)
			ctx = context.WithValue(ctx, "userRole", claims.Role)
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}

func AdminMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, ok := r.Context().Value("userRole").(string)
			if !ok || role != "admin" {
				responses.SendErrorResponse(w, http.StatusForbidden, "Access denied: admin role required")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
