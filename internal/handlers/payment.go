// internal/handlers/payment.go
package handlers

import (
    "database/sql"
    "encoding/json"
    "net/http"
    "time"

    "cctv-api/internal/models"
    "cctv-api/internal/responses"
)

func RequestPayment(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        userID, ok := r.Context().Value("userId").(int)
        if !ok {
            responses.SendErrorResponse(w, http.StatusUnauthorized, "Authentication required")
            return
        }

        // Generate payment data (in real app, integrate with payment gateway)
        paymentData := map[string]interface{}{
            "amount":      15000,
            "description": "Akses penuh CCTV seumur hidup",
            "payment_code": "CCTV-" + utils.GenerateRandomStringSimple(8),
            "expiry":      time.Now().Add(24 * time.Hour).Format(time.RFC3339),
            "qr_code_url": "https://api.qrserver.com/v1/create-qr-code/?size=150x150&data=CCTVPAY-" + utils.GenerateRandomStringSimple(16),
        }

        responses.SendSuccessResponse(w, http.StatusOK, paymentData)
    }
}

func ConfirmPayment(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        userID, ok := r.Context().Value("userId").(int)
        if !ok {
            responses.SendErrorResponse(w, http.StatusUnauthorized, "Authentication required")
            return
        }

        var req struct {
            PaymentCode string `json:"payment_code"`
        }

        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            responses.SendErrorResponse(w, http.StatusBadRequest, "Invalid request format")
            return
        }

        // In real app, verify payment with payment gateway
        // Here we just simulate successful payment

        now := time.Now()
        _, err := db.Exec(`
            UPDATE users 
            SET account_status = 'paid', payment_date = $1 
            WHERE id = $2
        `, now, userID)

        if err != nil {
            responses.SendErrorResponse(w, http.StatusInternalServerError, "Failed to update account status")
            return
        }

        responses.SendSuccessResponse(w, http.StatusOK, map[string]string{
            "message": "Pembayaran berhasil! Akun Anda sekarang memiliki akses penuh.",
        })
    }
}