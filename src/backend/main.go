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
	"AI-Note-Taker/queue"
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
	defer func() { _ = db.Close() }()

	if err = db.Ping(); err != nil {
		log.Println("failed to connect to database:", err)
	}

	if err = migrations.RunMigrations(os.Getenv("DATABASE_URL")); err != nil {
		log.Println(err)
	}

	log.Println("database connected succesfully")

	queue.StartWorkers(db)

	apiHandler := &api.Handler{DB: db}
	authHandler := &auth.Handler{DB: db}

	mainMux := http.NewServeMux()

	// api docs
	mainMux.Handle("GET /api-docs", http.HandlerFunc(api.DocsHandler))
	mainMux.Handle("GET /api-docs/openapi.yaml", http.HandlerFunc(api.OpenAPISpecHandler))

	// auth endpoints
	mainMux.Handle("POST /api/auth/register", http.HandlerFunc(authHandler.RegisterUserHandler))
	mainMux.Handle("POST /api/auth/login", http.HandlerFunc(authHandler.LoginHandler))
	mainMux.Handle("POST /api/auth/logout", http.HandlerFunc(auth.LogoutHandler))
	mainMux.Handle("GET /api/auth/me", http.HandlerFunc(auth.UserDataHandler))

	// upload endpoints
	mainMux.Handle("POST /api/uploads", middleware.AuthMiddleware(http.HandlerFunc(apiHandler.UploadHandler)))
	mainMux.Handle("GET /api/uploads", middleware.AuthMiddleware(http.HandlerFunc(apiHandler.GetUploadsHandler)))
	mainMux.Handle("DELETE /api/uploads/{id}", middleware.AuthMiddleware(http.HandlerFunc(apiHandler.DeleteUploadHandler)))
	mainMux.Handle("GET /api/uploads/{id}/notes", middleware.AuthMiddleware(http.HandlerFunc(apiHandler.GetNoteByUploadIDHandler)))
	mainMux.Handle("POST /api/uploads/{id}/notes/regenerate", middleware.AuthMiddleware(http.HandlerFunc(apiHandler.RegenerateNoteHandler)))

	// tag endpoints
	mainMux.Handle("GET /api/tags", middleware.AuthMiddleware(http.HandlerFunc(apiHandler.GetTagsHandler)))
	mainMux.Handle("POST /api/tags", middleware.AuthMiddleware(http.HandlerFunc(apiHandler.CreateTagHandler)))
	mainMux.Handle("DELETE /api/tags/{id}", middleware.AuthMiddleware(http.HandlerFunc(apiHandler.DeleteTagHandler)))
	mainMux.Handle("POST /api/uploads/{id}/tags", middleware.AuthMiddleware(http.HandlerFunc(apiHandler.AddTagToUploadHandler)))
	mainMux.Handle("DELETE /api/uploads/{id}/tags/{tagId}", middleware.AuthMiddleware(http.HandlerFunc(apiHandler.RemoveTagFromUploadHandler)))

	// static server for hosting on localhost:8080
	fs_ui := http.FileServer(http.Dir("../ui"))
	fs_frontend := http.FileServer(http.Dir("../frontend"))
	mainMux.Handle("/frontend/", http.StripPrefix("/frontend/", fs_frontend))
	mainMux.Handle("/", fs_ui)

	// enable cors
	fmt.Println("Server running on http://localhost:8080")
	fmt.Println("API docs available at http://localhost:8080/api-docs")
	log.Fatal(http.ListenAndServe(":8080", middleware.EnableCORS(mainMux)))
}
