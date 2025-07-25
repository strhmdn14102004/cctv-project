package handlers

import (
	"database/sql"
	"net/http"

	"cctv-api/internal/models"
	"cctv-api/internal/responses"
)

func UpgradeAccount(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value("userId").(int)

		_, err := db.Exec("UPDATE users SET account_status = 'paid' WHERE id = $1", userID)
		if err != nil {
			responses.SendErrorResponse(w, http.StatusInternalServerError, "Failed to upgrade account")
			return
		}

		responses.SendSuccessResponse(w, http.StatusOK, map[string]string{
			"message": "Account upgraded to paid successfully",
		})
	}
}