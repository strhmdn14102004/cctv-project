package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"cctv-api/internal/models"
	"cctv-api/internal/responses"
	"cctv-api/internal/services"
	"cctv-api/internal/utils"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/time/rate"
)

func Login(db *sql.DB, jwtUtil *utils.JWTUtil) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var creds struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}

		if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
			responses.SendErrorResponse(w, http.StatusBadRequest, "Invalid request format")
			return
		}

		if creds.Username == "" || creds.Password == "" {
			responses.SendErrorResponse(w, http.StatusBadRequest, "Username and password are required")
			return
		}

		var user models.User
		var lastLogin sql.NullTime
		var sessionToken sql.NullString
		err := db.QueryRow(`
			SELECT id, username, email, password, name, photo_url, role, account_status, 
			last_login, session_token
			FROM users WHERE username = $1 OR email = $1
		`, creds.Username).Scan(
			&user.ID, &user.Username, &user.Email, &user.Password,
			&user.Name, &user.PhotoURL, &user.Role, &user.AccountStatus,
			&lastLogin, &sessionToken,
		)

		if err != nil {
			if err == sql.ErrNoRows {
				responses.SendErrorResponse(w, http.StatusUnauthorized, "Invalid username or password")
			} else {
				responses.SendErrorResponse(w, http.StatusInternalServerError, "Database error")
			}
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(creds.Password)); err != nil {
			responses.SendErrorResponse(w, http.StatusUnauthorized, "Invalid username or password")
			return
		}

		// Check if user is already logged in
		if lastLogin.Valid && sessionToken.Valid {
			// Check if token is still valid
			_, err := jwtUtil.ValidateToken(sessionToken.String)
			if err == nil {
				responses.SendErrorResponse(w, http.StatusConflict, "User already logged in on another device")
				return
			}
		}

		token, err := jwtUtil.GenerateToken(user.ID, user.Role)
		if err != nil {
			responses.SendErrorResponse(w, http.StatusInternalServerError, "Failed to generate token")
			return
		}

		// Update last login and session token
		now := time.Now()
		_, err = db.Exec(`
			UPDATE users 
			SET last_login = $1, session_token = $2 
			WHERE id = $3
		`, now, token, user.ID)
		if err != nil {
			responses.SendErrorResponse(w, http.StatusInternalServerError, "Failed to update login status")
			return
		}

		userResponse := models.UserResponse{
			ID:            user.ID,
			Username:      user.Username,
			Email:         user.Email,
			Name:          user.Name,
			PhotoURL:      user.PhotoURL,
			Role:          user.Role,
			AccountStatus: user.AccountStatus,
		}

		responses.SendSuccessResponse(w, http.StatusOK, map[string]interface{}{
			"token": token,
			"user":  userResponse,
		})
	}
}

func Register(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var user struct {
			Username string  `json:"username" validate:"required"`
			Email    string  `json:"email" validate:"required,email"`
			Password string  `json:"password" validate:"required,min=8"`
			Name     string  `json:"name" validate:"required"`
			PhotoURL *string `json:"photoUrl"`
		}

		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			responses.SendErrorResponse(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
			return
		}

		// Validasi input
		if len(user.Password) < 8 {
			responses.SendErrorResponse(w, http.StatusBadRequest, "Password must be at least 8 characters")
			return
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			responses.SendErrorResponse(w, http.StatusInternalServerError, "Failed to hash password: "+err.Error())
			return
		}

		// Handle photo URL jika nil
		var photoUrl interface{}
		if user.PhotoURL != nil {
			photoUrl = *user.PhotoURL
		} else {
			photoUrl = nil
		}

		_, err = db.Exec(`
			INSERT INTO users (username, email, password, name, photo_url, role, account_status) 
			VALUES ($1, $2, $3, $4, $5, 'user', 'free')
		`, user.Username, user.Email, string(hashedPassword), user.Name, photoUrl)

		if err != nil {
			// Log error lengkap untuk debugging
			log.Printf("Database error during registration: %v", err)

			if err.Error() == `pq: duplicate key value violates unique constraint "users_username_key"` {
				responses.SendErrorResponse(w, http.StatusConflict, "Username already exists")
			} else if err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"` {
				responses.SendErrorResponse(w, http.StatusConflict, "Email already exists")
			} else {
				// Tampilkan error yang lebih spesifik
				responses.SendErrorResponse(w, http.StatusInternalServerError, "Failed to create user: "+err.Error())
			}
			return
		}

		responses.SendSuccessResponse(w, http.StatusCreated, map[string]string{
			"message": "User registered successfully",
		})
	}
}

func UpgradeAccount(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, ok := r.Context().Value(userClaimsKey).(*utils.Claims)
		if !ok {
			responses.SendErrorResponse(w, http.StatusUnauthorized, "Invalid user context")
			return
		}
		userID := claims.UserID

		result, err := db.Exec(`
            UPDATE users 
            SET account_status = 'paid', 
                fixed_cctv_ids = NULL, 
                updated_at = NOW()
            WHERE id = $1
        `, userID)

		if err != nil {
			log.Printf("Error upgrading account for user %d: %v", userID, err)
			responses.SendErrorResponse(w, http.StatusInternalServerError, "Failed to upgrade account")
			return
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			responses.SendErrorResponse(w, http.StatusNotFound, "User not found")
			return
		}

		responses.SendSuccessResponse(w, http.StatusOK, map[string]string{
			"message": "Account upgraded to paid successfully. You now have access to all CCTVs.",
		})
	}
}

func RateLimitMiddleware(limiter *rate.Limiter) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if !limiter.Allow() {
				responses.SendErrorResponse(w, http.StatusTooManyRequests, "Too many requests. Please try again later.")
				return
			}
			next(w, r)
		}
	}
}

func RequestDeviceReset(db *sql.DB, emailService *services.EmailService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.ResetRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			responses.SendErrorResponse(w, http.StatusBadRequest, "Invalid request format")
			return
		}

		// Check if email exists
		var userID int
		var username string
		err := db.QueryRow("SELECT id, username FROM users WHERE email = $1", req.Email).Scan(&userID, &username)
		if err != nil {
			if err == sql.ErrNoRows {
				responses.SendErrorResponse(w, http.StatusNotFound, "Email not found")
			} else {
				responses.SendErrorResponse(w, http.StatusInternalServerError, "Database error")
			}
			return
		}

		// Generate reset token
		resetToken := utils.GenerateRandomStringSimple(32)
		expiry := time.Now().Add(24 * time.Hour)

		_, err = db.Exec(`
			UPDATE users 
			SET reset_requested = true, reset_token = $1, reset_token_expiry = $2
			WHERE id = $3
		`, resetToken, expiry, userID)

		if err != nil {
			responses.SendErrorResponse(w, http.StatusInternalServerError, "Failed to process reset request")
			return
		}

		// Send email
		err = emailService.SendResetEmail(req.Email, resetToken)
		if err != nil {
			// Rollback the token if email fails
			db.Exec("UPDATE users SET reset_requested = false, reset_token = NULL, reset_token_expiry = NULL WHERE id = $1", userID)
			responses.SendErrorResponse(w, http.StatusInternalServerError, "Failed to send reset email")
			return
		}

		responses.SendSuccessResponse(w, http.StatusOK, map[string]string{
			"message": "Reset request submitted. Please check your email for instructions.",
		})
	}
}

func ConfirmDeviceReset(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.ResetPassword
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			responses.SendErrorResponse(w, http.StatusBadRequest, "Invalid request format")
			return
		}

		// Verify token and get user info
		var userID int
		var expiry time.Time
		var email string
		err := db.QueryRow(`
			SELECT id, reset_token_expiry, email 
			FROM users 
			WHERE reset_token = $1 AND reset_requested = true
		`, req.Token).Scan(&userID, &expiry, &email)

		if err != nil {
			if err == sql.ErrNoRows {
				responses.SendErrorResponse(w, http.StatusNotFound, "Invalid or expired reset token")
			} else {
				responses.SendErrorResponse(w, http.StatusInternalServerError, "Database error")
			}
			return
		}

		if time.Now().After(expiry) {
			responses.SendErrorResponse(w, http.StatusBadRequest, "Reset token has expired")
			return
		}

		// Verify password
		var hashedPassword string
		err = db.QueryRow("SELECT password FROM users WHERE id = $1", userID).Scan(&hashedPassword)
		if err != nil {
			responses.SendErrorResponse(w, http.StatusInternalServerError, "Database error")
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(req.Password)); err != nil {
			responses.SendErrorResponse(w, http.StatusUnauthorized, "Invalid password")
			return
		}

		// Clear device ID and reset flags
		_, err = db.Exec(`
			UPDATE users 
			SET device_id = NULL, reset_requested = false, reset_token = NULL, reset_token_expiry = NULL
			WHERE id = $1
		`, userID)

		if err != nil {
			responses.SendErrorResponse(w, http.StatusInternalServerError, "Failed to reset device")
			return
		}

		responses.SendSuccessResponse(w, http.StatusOK, map[string]string{
			"message": "Device reset successful. You can now login from a new device.",
		})
	}
}
func Logout(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, ok := r.Context().Value(userClaimsKey).(*utils.Claims)
		if !ok {
			responses.SendErrorResponse(w, http.StatusUnauthorized, "Invalid user context")
			return
		}
		userID := claims.UserID

		// Clear session token
		_, err := db.Exec(`
			UPDATE users 
			SET session_token = NULL 
			WHERE id = $1
		`, userID)
		if err != nil {
			responses.SendErrorResponse(w, http.StatusInternalServerError, "Failed to logout")
			return
		}

		responses.SendSuccessResponse(w, http.StatusOK, map[string]string{
			"message": "Logged out successfully",
		})
	}
}
