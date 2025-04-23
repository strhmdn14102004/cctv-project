package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"cctv-api/models"
	"cctv-api/responses"
	"cctv-api/utils"

	"golang.org/x/crypto/bcrypt"
)

func Login(db *sql.DB, jwtUtil *utils.JWTUtil) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var creds struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}

		if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
			responses.SendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		var user models.User
		err := db.QueryRow(`
			SELECT id, username, email, password, name, photo_url, role 
			FROM users WHERE username = $1
		`, creds.Username).Scan(
			&user.ID, &user.Username, &user.Email, &user.Password,
			&user.Name, &user.PhotoURL, &user.Role,
		)

		if err != nil {
			if err == sql.ErrNoRows {
				responses.SendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
			} else {
				responses.SendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
			}
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(creds.Password)); err != nil {
			responses.SendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		token, err := jwtUtil.GenerateToken(user.ID, user.Role)
		if err != nil {
			responses.SendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		// Prepare user response without sensitive data
		userResponse := models.UserResponse{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			Name:     user.Name,
			PhotoURL: user.PhotoURL,
			Role:     user.Role,
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
			responses.SendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		// Validate input
		if err := utils.Validate.Struct(user); err != nil {
			responses.SendValidationError(w, err)
			return
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			responses.SendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		_, err = db.Exec(`
			INSERT INTO users (username, email, password, name, photo_url, role) 
			VALUES ($1, $2, $3, $4, $5, 'user')
		`, user.Username, user.Email, string(hashedPassword), user.Name, user.PhotoURL)

		if err != nil {
			if err.Error() == `pq: duplicate key value violates unique constraint "users_username_key"` {
				responses.SendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
			} else if err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"` {
				responses.SendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
			} else {
				responses.SendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
			}
			return
		}

		responses.SendSuccessResponse(w, http.StatusCreated, map[string]string{
			"message": "User registered successfully",
		})
	}
}
