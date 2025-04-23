package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"cctv-api/models"
	"cctv-api/responses"
	"cctv-api/utils"

	"github.com/gorilla/mux"
)

func GetAllCCTVs(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get query parameters
		locationID := r.URL.Query().Get("locationId")
		isActive := r.URL.Query().Get("isActive")

		query := `
			SELECT 
				c.id, c.name, c.thumbnail_url, c.source_url, c.is_active, c.created_at, c.updated_at,
				l.id as location_id, l.name as location_name
			FROM cctvs c
			JOIN locations l ON c.location_id = l.id
			WHERE 1=1
		`
		args := []interface{}{}
		argPos := 1

		if locationID != "" {
			query += " AND l.id = $" + strconv.Itoa(argPos)
			args = append(args, locationID)
			argPos++
		}

		if isActive != "" {
			active, err := strconv.ParseBool(isActive)
			if err == nil {
				query += " AND c.is_active = $" + strconv.Itoa(argPos)
				args = append(args, active)
				argPos++
			}
		}

		query += " ORDER BY l.name ASC, c.name ASC"

		rows, err := db.Query(query, args...)
		if err != nil {
			responses.SendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
			return
		}
		defer rows.Close()

		var cctvs []models.CCTV
		for rows.Next() {
			var cctv models.CCTV
			var thumbnailUrl sql.NullString
			var loc models.Location

			err := rows.Scan(
				&cctv.ID,
				&cctv.Name,
				&thumbnailUrl,
				&cctv.SourceURL,
				&cctv.IsActive,
				&cctv.CreatedAt,
				&cctv.UpdatedAt,
				&loc.ID,
				&loc.Name,
			)
			if err != nil {
				responses.SendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
				return
			}

			if thumbnailUrl.Valid {
				cctv.ThumbnailURL = &thumbnailUrl.String
			}
			cctv.Location = &loc

			cctvs = append(cctvs, cctv)
		}

		responses.SendSuccessResponse(w, http.StatusOK, cctvs)
	}
}

func GetCCTVByID(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id, err := strconv.Atoi(vars["id"])
		if err != nil {
			responses.SendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		var cctv models.CCTV
		var thumbnailUrl sql.NullString
		var loc models.Location

		err = db.QueryRow(`
			SELECT 
				c.id, c.name, c.thumbnail_url, c.source_url, c.is_active, c.created_at, c.updated_at,
				l.id as location_id, l.name as location_name
			FROM cctvs c
			JOIN locations l ON c.location_id = l.id
			WHERE c.id = $1
		`, id).Scan(
			&cctv.ID,
			&cctv.Name,
			&thumbnailUrl,
			&cctv.SourceURL,
			&cctv.IsActive,
			&cctv.CreatedAt,
			&cctv.UpdatedAt,
			&loc.ID,
			&loc.Name,
		)

		if err != nil {
			if err == sql.ErrNoRows {
				responses.SendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
			} else {
				responses.SendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
			}
			return
		}

		if thumbnailUrl.Valid {
			cctv.ThumbnailURL = &thumbnailUrl.String
		}
		cctv.Location = &loc

		responses.SendSuccessResponse(w, http.StatusOK, cctv)
	}
}

func CreateCCTV(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.CreateCCTVRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			responses.SendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		// Validate input
		if err := utils.Validate.Struct(req); err != nil {
			responses.SendValidationError(w, err)
			return
		}

		var id int
		var thumbnailUrl interface{}
		if req.ThumbnailURL != nil {
			thumbnailUrl = *req.ThumbnailURL
		} else {
			thumbnailUrl = nil
		}

		err := db.QueryRow(`
			INSERT INTO cctvs (location_id, name, thumbnail_url, source_url)
			VALUES ($1, $2, $3, $4)
			RETURNING id
		`, req.LocationID, req.Name, thumbnailUrl, req.SourceURL).Scan(&id)

		if err != nil {
			if err.Error() == `pq: insert or update on table "cctvs" violates foreign key constraint "cctvs_location_id_fkey"` {
				responses.SendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
			} else {
				responses.SendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
			}
			return
		}

		responses.SendSuccessResponse(w, http.StatusCreated, map[string]int{"id": id})
	}
}

func UpdateCCTV(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id, err := strconv.Atoi(vars["id"])
		if err != nil {
			responses.SendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		var req models.UpdateCCTVRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			responses.SendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		// Validate input
		if err := utils.Validate.Struct(req); err != nil {
			responses.SendValidationError(w, err)
			return
		}

		// Build dynamic update query
		query := "UPDATE cctvs SET updated_at = NOW()"
		args := []interface{}{}
		argPos := 1

		if req.LocationID != nil {
			query += ", location_id = $" + strconv.Itoa(argPos)
			args = append(args, *req.LocationID)
			argPos++
		}

		if req.Name != nil {
			query += ", name = $" + strconv.Itoa(argPos)
			args = append(args, *req.Name)
			argPos++
		}

		if req.ThumbnailURL != nil {
			query += ", thumbnail_url = $" + strconv.Itoa(argPos)
			args = append(args, *req.ThumbnailURL)
			argPos++
		}

		if req.SourceURL != nil {
			query += ", source_url = $" + strconv.Itoa(argPos)
			args = append(args, *req.SourceURL)
			argPos++
		}

		if req.IsActive != nil {
			query += ", is_active = $" + strconv.Itoa(argPos)
			args = append(args, *req.IsActive)
			argPos++
		}

		query += " WHERE id = $" + strconv.Itoa(argPos)
		args = append(args, id)
		argPos++

		_, err = db.Exec(query, args...)
		if err != nil {
			if err.Error() == `pq: insert or update on table "cctvs" violates foreign key constraint "cctvs_location_id_fkey"` {
				responses.SendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
			} else {
				responses.SendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
			}
			return
		}

		responses.SendSuccessResponse(w, http.StatusOK, map[string]string{
			"message": "CCTV updated successfully",
		})
	}
}

func DeleteCCTV(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id, err := strconv.Atoi(vars["id"])
		if err != nil {
			responses.SendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		_, err = db.Exec("DELETE FROM cctvs WHERE id = $1", id)
		if err != nil {
			responses.SendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		responses.SendSuccessResponse(w, http.StatusOK, map[string]string{
			"message": "CCTV deleted successfully",
		})
	}
}
