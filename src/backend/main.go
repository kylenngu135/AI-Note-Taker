package main

import (
    "database/sql"
    "log"
    "os"
	"fmt"
	"net/http"

    _ "github.com/jackc/pgx/v5/stdlib"
    "github.com/joho/godotenv"

	"AI-Note-Taker/api"
	"AI-Note-Taker/middleware"
	"AI-Note-Taker/migrations"
	"AI-Note-Taker/storage"
)

func main() {
	if err := godotenv.Load("../../.env"); err != nil {
		log.Fatal("error loading .env file")
	}

	// R2 storage
	err := storage.InitR2(
		os.Getenv("R2_ACCOUNT_ID"),
		os.Getenv("R2_ACCESS_KEY_ID"),
		os.Getenv("R2_SECRET_ACCESS_KEY"),
	)
	if err != nil {
		log.Fatal(err)
	}

	db, err := sql.Open("pgx", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal("failed to open database:", err)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		log.Fatal("failed to connect to database:", err)
	}

    if err = migrations.RunMigrations(os.Getenv("DATABASE_URL")); err != nil {
        log.Fatal(err)
    }

	log.Println("database connected succesfully")

	api := &api.Handler{DB: db}

	// static server for hosting on localhost:8080
	fs_ui := http.FileServer(http.Dir("../ui"))
	fs_frontend := http.FileServer(http.Dir("../frontend"))
	fs_storage := http.FileServer(http.Dir("../../storage/media/"))
	http.Handle("/", fs_ui)
	http.Handle("/frontend/", http.StripPrefix("/frontend/", fs_frontend))
	http.Handle("/storage/", http.StripPrefix("/storage/", fs_storage))

	// api endpoints
	http.HandleFunc("/api/uploads/documents", api.DocumentUploadHandler)
	http.HandleFunc("/api/uploads/videos", api.VideoUploadHandler)
	http.HandleFunc("/api/uploads/audios", api.AudioUploadHandler)
	http.HandleFunc("/api/uploads", api.GetUploadsHandler)
	http.HandleFunc("/api/uploads/{id}", api.DeleteUploadHandler)

	// enable cors
	fmt.Println("Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", middleware.EnableCORS(http.DefaultServeMux)))
}
