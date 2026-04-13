package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"

	"AI-Note-Taker/api"
	"AI-Note-Taker/auth"
	"AI-Note-Taker/middleware"
	"AI-Note-Taker/migrations"
	"AI-Note-Taker/storage"
)

func main() {
	if err := godotenv.Load("../../.env"); err != nil {
		log.Println("no .env file found, using environment variables")
	}

	// R2 storage
	err := storage.InitR2(
		os.Getenv("R2_ACCOUNT_ID"),
		os.Getenv("R2_ACCESS_KEY_ID"),
		os.Getenv("R2_SECRET_ACCESS_KEY"),
	)
	if err != nil {
		log.Println(err)
	}

	db, err := sql.Open("pgx", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Println("failed to open database:", err)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		log.Println("failed to connect to database:", err)
	}

	if err = migrations.RunMigrations(os.Getenv("DATABASE_URL")); err != nil {
		log.Println(err)
	}

	log.Println("database connected succesfully")

	api := &api.Handler{DB: db}
	authHandler := &auth.Handler{DB: db}

	mainMux := http.NewServeMux()

	// auth endpoints
	mainMux.Handle("POST /api/auth/register", http.HandlerFunc(authHandler.RegisterUserHandler))
	mainMux.Handle("POST /api/auth/login", http.HandlerFunc(authHandler.LoginHandler))
	mainMux.Handle("POST /api/auth/logout", http.HandlerFunc(auth.LogoutHandler))
	mainMux.Handle("GET /api/auth/me", http.HandlerFunc(auth.UserDataHandler))

	// upload endpoints
	mainMux.Handle("POST /api/uploads", middleware.AuthMiddleware(http.HandlerFunc(api.UploadHandler)))
	mainMux.Handle("GET /api/uploads", middleware.AuthMiddleware(http.HandlerFunc(api.GetUploadsHandler)))
	mainMux.Handle("DELETE /api/uploads/{id}", middleware.AuthMiddleware(http.HandlerFunc(api.DeleteUploadHandler)))
	mainMux.Handle("GET /api/uploads/{id}/notes", middleware.AuthMiddleware(http.HandlerFunc(api.GetNoteByUploadIDHandler)))
	mainMux.Handle("POST /api/uploads/{id}/notes/regenerate", middleware.AuthMiddleware(http.HandlerFunc(api.RegenerateNoteHandler)))

	// static server for hosting on localhost:8080
	fs_ui := http.FileServer(http.Dir("../ui"))
	fs_frontend := http.FileServer(http.Dir("../frontend"))
	mainMux.Handle("/frontend/", http.StripPrefix("/frontend/", fs_frontend))
	mainMux.Handle("/", fs_ui)

	// enable cors
	fmt.Println("Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", middleware.EnableCORS(mainMux)))
}
