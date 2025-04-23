package handlers

import (
	"context"
	"net/http"
	"strings"

	"cctv-api/responses"
	"cctv-api/utils"
)

// JWTMiddleware validates JWT tokens and sets user info in context
func JWTMiddleware(jwtUtil *utils.JWTUtil) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

			// Add claims to context
			ctx := context.WithValue(r.Context(), "userId", claims.UserID)
			ctx = context.WithValue(ctx, "userRole", claims.Role)
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}

// DeveloperMiddleware restricts access to developers only
func DeveloperMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, ok := r.Context().Value("userRole").(string)
			if !ok || role != "developer" {
				responses.SendErrorResponse(w, http.StatusForbidden, "Access denied: developer role required")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// AdminMiddleware restricts access to admins only (optional)
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
