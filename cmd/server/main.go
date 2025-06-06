package main

import (
	"cctv-api/internal/config"
	"cctv-api/internal/database"
	"cctv-api/internal/handlers"
	"cctv-api/internal/utils"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Initialize database
	db, err := database.NewDatabase(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Initialize JWT utility
	jwtUtil := utils.NewJWTUtil(cfg.JWTSecret, cfg.JWTExpiration)

	// Create router
	router := mux.NewRouter()

	// Health check endpoint
	router.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	// Auth routes
	authRouter := router.PathPrefix("/api/auth").Subrouter()
	{
		authRouter.HandleFunc("/login", handlers.Login(db.DB, jwtUtil)).Methods("POST")
		authRouter.HandleFunc("/register", handlers.Register(db.DB)).Methods("POST")
	}

	// Public routes
	publicRouter := router.PathPrefix("/api/public").Subrouter()
	{
		publicRouter.HandleFunc("/locations", handlers.GetAllLocations(db.DB)).Methods("GET")
		publicRouter.HandleFunc("/cctvs", handlers.GetAllCCTVs(db.DB)).Methods("GET")
		publicRouter.HandleFunc("/cctvs/{id:[0-9]+}", handlers.GetCCTVByID(db.DB)).Methods("GET")
	}

	// Authenticated routes
	apiRouter := router.PathPrefix("/api").Subrouter()
	apiRouter.Use(handlers.JWTMiddleware(jwtUtil))
	{
		// Locations
		apiRouter.HandleFunc("/locations", handlers.CreateLocation(db.DB)).Methods("POST")
		apiRouter.HandleFunc("/locations/{id:[0-9]+}", handlers.DeleteLocation(db.DB)).Methods("DELETE")

		// CCTVs
		apiRouter.HandleFunc("/cctvs", handlers.CreateCCTV(db.DB)).Methods("POST")
		apiRouter.HandleFunc("/cctvs/{id:[0-9]+}", handlers.UpdateCCTV(db.DB)).Methods("PUT")
		apiRouter.HandleFunc("/cctvs/{id:[0-9]+}", handlers.DeleteCCTV(db.DB)).Methods("DELETE")
	}

	// CORS configuration
	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	})

	// Start server
	log.Printf("Server running on port %s", cfg.AppPort)
	log.Fatal(http.ListenAndServe(":"+cfg.AppPort, corsHandler.Handler(router)))
}
