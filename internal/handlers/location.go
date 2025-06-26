package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"cctv-api/internal/models"
	"cctv-api/internal/responses"

	"github.com/gorilla/mux"
)

func GetAllLocations(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query(`
			SELECT id, name, created_at, updated_at 
			FROM locations 
			ORDER BY name ASC
		`)
		if err != nil {
			responses.SendErrorResponse(w, http.StatusInternalServerError, "Failed to fetch locations")
			return
		}
		defer rows.Close()

		var locations []models.Location
		for rows.Next() {
			var loc models.Location
			err := rows.Scan(&loc.ID, &loc.Name, &loc.CreatedAt, &loc.UpdatedAt)
			if err != nil {
				responses.SendErrorResponse(w, http.StatusInternalServerError, "Failed to scan location data")
				return
			}
			locations = append(locations, loc)
		}

		responses.SendSuccessResponse(w, http.StatusOK, locations)
	}
}

func CreateLocation(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var loc struct {
			Name string `json:"name" validate:"required"`
		}

		if err := json.NewDecoder(r.Body).Decode(&loc); err != nil {
			responses.SendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		// Check for existing location with same name
		var existingID int
		err := db.QueryRow("SELECT id FROM locations WHERE name = $1", loc.Name).Scan(&existingID)
		if err == nil {
			responses.SendErrorResponse(w, http.StatusConflict,
				"Location with name '"+loc.Name+"' already exists (ID: "+strconv.Itoa(existingID)+")")
			return
		} else if err != sql.ErrNoRows {
			responses.SendErrorResponse(w, http.StatusInternalServerError, "Failed to check for duplicate location")
			return
		}

		var id int
		err = db.QueryRow(`
			INSERT INTO locations (name) 
			VALUES ($1) 
			RETURNING id
		`, loc.Name).Scan(&id)

		if err != nil {
			responses.SendErrorResponse(w, http.StatusInternalServerError, "Failed to create location")
			return
		}

		responses.SendSuccessResponse(w, http.StatusCreated, map[string]int{"id": id})
	}
}

func DeleteLocation(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id, err := strconv.Atoi(vars["id"])
		if err != nil {
			responses.SendErrorResponse(w, http.StatusBadRequest, "Invalid location ID")
			return
		}

		// Check if location is used by any CCTV
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM cctvs WHERE location_id = $1", id).Scan(&count)
		if err != nil {
			responses.SendErrorResponse(w, http.StatusInternalServerError, "Failed to check location usage")
			return
		}

		if count > 0 {
			responses.SendErrorResponse(w, http.StatusConflict, "Cannot delete location with associated CCTV")
			return
		}

		result, err := db.Exec("DELETE FROM locations WHERE id = $1", id)
		if err != nil {
			responses.SendErrorResponse(w, http.StatusInternalServerError, "Failed to delete location")
			return
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			responses.SendErrorResponse(w, http.StatusNotFound, "Location not found")
			return
		}

		responses.SendSuccessResponse(w, http.StatusOK, map[string]string{
			"message": "Location deleted successfully",
		})
	}
}
