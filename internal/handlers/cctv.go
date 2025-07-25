package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"log"
	"cctv-api/internal/models"
	"cctv-api/internal/responses"
	"cctv-api/internal/utils"

	"github.com/gorilla/mux"
	"github.com/lib/pq"
)

func GetAllCCTVs(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        claims, ok := r.Context().Value(userClaimsKey).(*utils.Claims)
        if !ok {
            responses.SendErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
            return
        }
        userID := claims.UserID

        // Ambil account_status dan fixed_cctv_ids
        var accountStatus string
        var fixedIDs []int
        var fixedIDsRaw sql.NullString

        err := db.QueryRow(`SELECT account_status, fixed_cctv_ids FROM users WHERE id = $1`, userID).
            Scan(&accountStatus, &fixedIDsRaw)
        if err != nil {
            responses.SendErrorResponse(w, http.StatusInternalServerError, "Failed to get user info")
            return
        }

        // Decode fixed_cctv_ids jika ada
        if fixedIDsRaw.Valid {
            _ = json.Unmarshal([]byte(fixedIDsRaw.String), &fixedIDs)
        }

        // Jika akun free dan belum punya fixed_cctv_ids → generate dan simpan
        if accountStatus == "free" && len(fixedIDs) == 0 {
            rows, err := db.Query(`SELECT id FROM cctvs WHERE is_active = true ORDER BY RANDOM() LIMIT 10`)
            if err != nil {
                responses.SendErrorResponse(w, http.StatusInternalServerError, "Failed to fetch random CCTVs")
                return
            }
            defer rows.Close()

            for rows.Next() {
                var id int
                if err := rows.Scan(&id); err != nil {
                    responses.SendErrorResponse(w, http.StatusInternalServerError, "Failed to scan CCTV ID")
                    return
                }
                fixedIDs = append(fixedIDs, id)
            }

            if len(fixedIDs) > 0 {
                fixedJSON, _ := json.Marshal(fixedIDs)
                _, err = db.Exec(`UPDATE users SET fixed_cctv_ids = $1 WHERE id = $2`, string(fixedJSON), userID)
                if err != nil {
                    log.Printf("Failed to save fixed_cctv_ids: %v", err)
                }
            }
        }

        // Base query
        query := `
            SELECT 
                c.id, c.name, c.thumbnail_url, c.source_url, c.is_active, c.created_at, c.updated_at,
                l.id as location_id, l.name as location_name
            FROM cctvs c
            JOIN locations l ON c.location_id = l.id
            WHERE c.is_active = true
        `
        args := []interface{}{}
        argPos := 1

        // Filter untuk akun free
        if accountStatus == "free" {
            if len(fixedIDs) == 0 {
                responses.SendSuccessResponse(w, http.StatusOK, []models.CCTV{})
                return
            }

            query += " AND c.id = ANY($" + strconv.Itoa(argPos) + ")"
            args = append(args, pq.Array(fixedIDs))
            argPos++
        }

        // Filter locationID
        locationID := r.URL.Query().Get("locationId")
        if locationID != "" {
            query += " AND l.id = $" + strconv.Itoa(argPos)
            args = append(args, locationID)
            argPos++
        }

        query += " ORDER BY l.name ASC, c.name ASC"

        // Eksekusi dan scan data
        rows, err := db.Query(query, args...)
        if err != nil {
            responses.SendErrorResponse(w, http.StatusInternalServerError, "Failed to fetch CCTVs: "+err.Error())
            return
        }
        defer rows.Close()

        var cctvs []models.CCTV
        for rows.Next() {
            var cctv models.CCTV
            var thumbnail sql.NullString
            var loc models.Location
            err := rows.Scan(&cctv.ID, &cctv.Name, &thumbnail, &cctv.SourceURL, &cctv.IsActive, 
                &cctv.CreatedAt, &cctv.UpdatedAt, &loc.ID, &loc.Name)
            if err != nil {
                responses.SendErrorResponse(w, http.StatusInternalServerError, "Failed to scan CCTV data")
                return
            }
            if thumbnail.Valid {
                cctv.ThumbnailURL = &thumbnail.String
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
			responses.SendErrorResponse(w, http.StatusBadRequest, "Invalid CCTV ID")
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
				responses.SendErrorResponse(w, http.StatusNotFound, "CCTV not found")
			} else {
				responses.SendErrorResponse(w, http.StatusInternalServerError, "Failed to fetch CCTV")
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

		if err := utils.Validate.Struct(req); err != nil {
			responses.SendValidationError(w, err)
			return
		}

		// Check for duplicate name
		var existingID int
		err := db.QueryRow("SELECT id FROM cctvs WHERE name = $1", req.Name).Scan(&existingID)
		if err == nil {
			responses.SendErrorResponse(w, http.StatusConflict,
				"CCTV with name '"+req.Name+"' already exists (ID: "+strconv.Itoa(existingID)+")")
			return
		} else if err != sql.ErrNoRows {
			responses.SendErrorResponse(w, http.StatusInternalServerError, "Failed to check for duplicate name")
			return
		}

		// Check for duplicate source URL
		err = db.QueryRow("SELECT id, name FROM cctvs WHERE source_url = $1", req.SourceURL).Scan(&existingID, &req.Name)
		if err == nil {
			responses.SendErrorResponse(w, http.StatusConflict,
				"CCTV with this source URL already exists (ID: "+strconv.Itoa(existingID)+", Name: '"+req.Name+"')")
			return
		} else if err != sql.ErrNoRows {
			responses.SendErrorResponse(w, http.StatusInternalServerError, "Failed to check for duplicate source URL")
			return
		}

		var thumbnailUrl interface{}
		if req.ThumbnailURL != nil {
			thumbnailUrl = *req.ThumbnailURL
		} else {
			thumbnailUrl = nil
		}

		var id int
		err = db.QueryRow(`
			INSERT INTO cctvs (location_id, name, thumbnail_url, source_url)
			VALUES ($1, $2, $3, $4)
			RETURNING id
		`, req.LocationID, req.Name, thumbnailUrl, req.SourceURL).Scan(&id)

		if err != nil {
			if err.Error() == `pq: insert or update on table "cctvs" violates foreign key constraint "cctvs_location_id_fkey"` {
				responses.SendErrorResponse(w, http.StatusBadRequest, "Invalid location ID")
			} else {
				responses.SendErrorResponse(w, http.StatusInternalServerError, "Failed to create CCTV")
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
			responses.SendErrorResponse(w, http.StatusBadRequest, "Invalid CCTV ID")
			return
		}

		var req models.UpdateCCTVRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			responses.SendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		if err := utils.Validate.Struct(req); err != nil {
			responses.SendValidationError(w, err)
			return
		}

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

		result, err := db.Exec(query, args...)
		if err != nil {
			if err.Error() == `pq: insert or update on table "cctvs" violates foreign key constraint "cctvs_location_id_fkey"` {
				responses.SendErrorResponse(w, http.StatusBadRequest, "Invalid location ID")
			} else {
				responses.SendErrorResponse(w, http.StatusInternalServerError, "Failed to update CCTV")
			}
			return
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			responses.SendErrorResponse(w, http.StatusNotFound, "CCTV not found")
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
			responses.SendErrorResponse(w, http.StatusBadRequest, "Invalid CCTV ID")
			return
		}

		result, err := db.Exec("DELETE FROM cctvs WHERE id = $1", id)
		if err != nil {
			responses.SendErrorResponse(w, http.StatusInternalServerError, "Failed to delete CCTV")
			return
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			responses.SendErrorResponse(w, http.StatusNotFound, "CCTV not found")
			return
		}

		responses.SendSuccessResponse(w, http.StatusOK, map[string]string{
			"message": "CCTV deleted successfully",
		})
	}
}
