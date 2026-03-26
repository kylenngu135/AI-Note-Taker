package auth

import (
	"AI-Note-Taker/api"
	"database/sql"
	"encoding/json"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"net/mail"
	"os"
	"time"
	"fmt"
)

const COMP_COST = 12

type Handler struct {
	DB *sql.DB
}

type User struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *Handler) RegisterUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// grab user data
	var newUser User
	err := json.NewDecoder(r.Body).Decode(&newUser)

	// validate the email
	if !isValidEmail(newUser.Email) {
		http.Error(w, "invalid email address", http.StatusBadRequest)
		return
	}

	// check if email is already registered
	exists, err := api.CheckUserExists(h.DB, newUser.Email)
	if err != nil {
		http.Error(w, "failed to query database", http.StatusInternalServerError)
	}
	if exists {
		http.Error(w, "email already exists", http.StatusConflict)
		return
	}

	// encrypt password with bcrypt
	bytePassword := []byte(newUser.Password)
	hashedBytePassword, err := bcrypt.GenerateFromPassword(bytePassword, COMP_COST)
	if err != nil {
		http.Error(w, "failed to hash password", http.StatusInternalServerError)
		return
	}
	hashedPassword := string(hashedBytePassword)

	// store in db
	userID := uuid.New().String()
	err = api.InsertUser(h.DB, userID, newUser.Email, hashedPassword)
	if err != nil {
		http.Error(w, "failed to insert into database", http.StatusInternalServerError)
	}

	// create JWT
	jwtToken, err := generateJWT(userID, newUser.Email)
	if err != nil {
		http.Error(w, "failed to generate token", http.StatusInternalServerError)
	}

	// send response
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"token": jwtToken,
	})

	return
}

/*
func(h *Handler) LoginHandler(w http.ResponseWriter, r *http.Request) {

}

func(h *Handler) LogoutHandler(w http.ResponseWriter, r *http.Request) {

}

func (h *Handler) GetUserDataHandler(w http.ResponseWriter, r *http.Request) {

}
*/

func isValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

func generateJWT(userID, email string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"exo":     time.Now().Add(time.Hour * 24).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signed, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return signed, nil
}
