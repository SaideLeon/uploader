package handlers

import (
	"encoding/json"
	"net/http"

	"gorm.io/gorm"
	"github.com/GoogleCloudPlatform/golang-samples/run/helloworld/models"
	"github.com/GoogleCloudPlatform/golang-samples/run/helloworld/util"
)

type AuthRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Message     string       `json:"message"`
	Token       string       `json:"token,omitempty"`
	User        *models.User `json:"user,omitempty"`
	ForgeAPIKey string       `json:"forge_api_key,omitempty"`
}

// RegisterHandler godoc
// @Summary Register a new user
// @Description Creates a new user account and returns the user info along with an API key.
// @Tags auth
// @Accept  json
// @Produce  json
// @Param   auth_request  body  AuthRequest  true  "User registration details"
// @Success 201 {object} AuthResponse "User created successfully"
// @Failure 400 {string} string "Invalid request body or missing fields"
// @Failure 500 {string} string "Could not create user"
// @Router /register [post]
func RegisterHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var req AuthRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Validação básica
		if req.Email == "" || req.Password == "" {
			http.Error(w, "Email and password are required", http.StatusBadRequest)
			return
		}

		user := &models.User{
			Email:    req.Email,
			Password: req.Password,
		}

		// O hook BeforeCreate irá gerar a API key e hashear a senha
		if err := db.Create(user).Error; err != nil {
			http.Error(w, "Could not create user: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Não retornar a senha
		user.Password = ""

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(AuthResponse{
			Message:     "User created successfully",
			User:        user,
			ForgeAPIKey: user.ForgeAPIKey,
		})
	}
}

// LoginHandler godoc
// @Summary Log in a user
// @Description Authenticates a user and returns a JWT token.
// @Tags auth
// @Accept  json
// @Produce  json
// @Param   auth_request  body  AuthRequest  true  "User login credentials"
// @Success 200 {object} AuthResponse "Logged in successfully"
// @Failure 400 {string} string "Invalid request body"
// @Failure 401 {string} string "Invalid credentials"
// @Failure 500 {string} string "Could not generate token"
// @Router /login [post]
func LoginHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var req AuthRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		var user models.User
		if err := db.First(&user, "email = ?", req.Email).Error; err != nil {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}

		if !user.CheckPassword(req.Password) {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}

		token, err := util.GenerateJWT(&user)
		if err != nil {
			http.Error(w, "Could not generate token", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(AuthResponse{
			Message: "Logged in successfully",
			Token:   token,
		})
	}
}
