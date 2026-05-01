package auth

import (
	"AI-Note-Taker/api"
	"AI-Note-Taker/middleware"
	"database/sql"
	"encoding/json"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"net/mail"
	"os"
	"time"
	"log"
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
	if err != nil {
		http.Error(w, "failed to decode input", http.StatusInternalServerError)
	}

	// validate the email
	if !isValidEmail(newUser.Email) {
		http.Error(w, "invalid email address", http.StatusBadRequest)
		return
	}

	// check if email is already registered
	exists, err := api.CheckUserExists(h.DB, newUser.Email)
	if err != nil {
		http.Error(w, "failed to query database", http.StatusInternalServerError)
		return
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
		return
	}

	// create JWT
	jwtToken, err := generateJWT(userID, newUser.Email)
	if err != nil {
		http.Error(w, "failed to generate token", http.StatusInternalServerError)
		return
	}

	createCookie(w, jwtToken)

	// send response
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"token": jwtToken,
	})

	return
}

func (h *Handler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// grab user data
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "failed to decode input", http.StatusInternalServerError)
	}


	//TODO:Redundant Code, check to see if you can remove

	// validate the email
	if !isValidEmail(user.Email) {
		http.Error(w, "invalid email address", http.StatusBadRequest)
		return
	}

	hashedPassword, err := api.GetHashedPasswordByEmail(h.DB, user.Email)
	if err == sql.ErrNoRows {
		log.Printf("%v", err)
		http.Error(w, "invalid email or password", http.StatusBadRequest)
		return
	}
	if err != nil {
		log.Printf("%v", err)
		http.Error(w, "failed to query database", http.StatusInternalServerError)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(user.Password))
	if err != nil {
		log.Printf("%v", err)
		http.Error(w, "invalid email or password", http.StatusBadRequest)
		return
	}

	// TODO:Fix it so you only query the database once

	userID, err := api.GetUserIDByEmail(h.DB, user.Email)
	if err != nil {
		log.Printf("%v", err)
		http.Error(w, "failed to query database", http.StatusInternalServerError)
		return
	}

	jwtToken, err := generateJWT(userID, user.Email)
	if err != nil {
		log.Printf("%v", err)
		http.Error(w, "failed to generate token", http.StatusInternalServerError)
		return
	}

	createCookie(w, jwtToken)

	// send response
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusCreated)

	// TODO: Change what we return in the body

	json.NewEncoder(w).Encode(map[string]string{
		"token": jwtToken,
	})

	return
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
    http.SetCookie(w, &http.Cookie{
        Name:     "auth_token",
        Value:    "",
        MaxAge:   -1,
        Path:     "/",
        HttpOnly: true,
        Secure:   true,
        SameSite: http.SameSiteStrictMode,
    })

    w.WriteHeader(http.StatusOK)

	return
}

func UserDataHandler(w http.ResponseWriter, r *http.Request) {
    cookie, err := r.Cookie("auth_token")
    if err != nil {
        http.Error(w, "unauthorized", http.StatusUnauthorized)
        return
    }

    claims, err := middleware.ValidateJWT(cookie.Value)
    if err != nil {
        http.Error(w, "invalid token", http.StatusUnauthorized)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{
        "user_id": (*claims)["user_id"].(string),
        "email":   (*claims)["email"].(string),
    })
}

func isValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

func generateJWT(userID, email string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return signed, nil
}

func createCookie(w http.ResponseWriter, jwtToken string) {
	http.SetCookie(w, &http.Cookie{
		Name: 		"auth_token",
		Value: 		jwtToken,
		HttpOnly: 	true,
		Secure: 	true,
		SameSite:	http.SameSiteStrictMode,
		Path:		"/",
		MaxAge:		86400,
	})

	return
}
