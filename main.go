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
	cfg := config.LoadConfig()
	db, err := database.InitDB(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	jwtUtil := utils.NewJWTUtil(cfg.JWTSecret, cfg.JWTExpiration)
	router := mux.NewRouter()

	// Health Check
	router.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	}).Methods("GET")

	// Auth Routes
	authRouter := router.PathPrefix("/api/auth").Subrouter()
	{
		authRouter.HandleFunc("/login", handlers.Login(db, jwtUtil)).Methods("POST")
		authRouter.HandleFunc("/register", handlers.Register(db)).Methods("POST")
	}

	// Public Routes
	publicRouter := router.PathPrefix("/api/public").Subrouter()
	{
		publicRouter.HandleFunc("/locations", handlers.GetAllLocations(db)).Methods("GET")
		publicRouter.HandleFunc("/cctvs", handlers.GetAllCCTVs(db)).Methods("GET")
		publicRouter.HandleFunc("/cctvs/{id}", handlers.GetCCTVByID(db)).Methods("GET")
	}

	// Authenticated Routes
	apiRouter := router.PathPrefix("/api").Subrouter()
	apiRouter.Use(handlers.JWTMiddleware(jwtUtil))
	{
		// Locations
		apiRouter.HandleFunc("/locations", handlers.CreateLocation(db)).Methods("POST")
		apiRouter.HandleFunc("/locations/{id}", handlers.DeleteLocation(db)).Methods("DELETE")

		// CCTVs
		apiRouter.HandleFunc("/cctvs", handlers.CreateCCTV(db)).Methods("POST")
		apiRouter.HandleFunc("/cctvs/{id}", handlers.UpdateCCTV(db)).Methods("PUT")
		apiRouter.HandleFunc("/cctvs/{id}", handlers.DeleteCCTV(db)).Methods("DELETE")
	}

	// CORS Configuration
	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	})

	log.Printf("Server running on port %s", cfg.AppPort)
	log.Fatal(http.ListenAndServe(":"+cfg.AppPort, corsHandler.Handler(router)))
}
