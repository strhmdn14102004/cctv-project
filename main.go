package main

import (
	"cctv-api/config"
	"cctv-api/database"
	"cctv-api/handlers"
	"cctv-api/utils"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Initialize database
	db, err := database.InitDB(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Initialize JWT
	jwtUtil := utils.NewJWTUtil(cfg.JWTSecret, cfg.JWTExpiration)

	// Create router
	router := mux.NewRouter()

	// Auth routes
	router.HandleFunc("/api/login", handlers.Login(db, jwtUtil)).Methods("POST")
	router.HandleFunc("/api/register", handlers.Register(db)).Methods("POST")

	// Public routes
	publicRouter := router.PathPrefix("/api/public").Subrouter()
	publicRouter.HandleFunc("/locations", handlers.GetAllLocations(db)).Methods("GET")

	// Authenticated routes
	authRouter := router.PathPrefix("/api").Subrouter()
	authRouter.Use(handlers.JWTMiddleware(jwtUtil))

	// CCTV routes
	authRouter.HandleFunc("/cctvs", handlers.GetAllCCTVs(db)).Methods("GET")
	authRouter.HandleFunc("/cctvs/{id}", handlers.GetCCTVByID(db)).Methods("GET")

	// Location routes
	authRouter.HandleFunc("/locations", handlers.CreateLocation(db)).Methods("POST")
	authRouter.HandleFunc("/locations/{id}", handlers.DeleteLocation(db)).Methods("DELETE")

	// Developer-only routes
	devRouter := router.PathPrefix("/api").Subrouter()
	devRouter.Use(handlers.JWTMiddleware(jwtUtil))
	devRouter.Use(handlers.DeveloperMiddleware())
	devRouter.HandleFunc("/cctvs", handlers.CreateCCTV(db)).Methods("POST")
	devRouter.HandleFunc("/cctvs/{id}", handlers.UpdateCCTV(db)).Methods("PUT")
	devRouter.HandleFunc("/cctvs/{id}", handlers.DeleteCCTV(db)).Methods("DELETE")

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
