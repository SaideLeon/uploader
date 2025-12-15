package handlers

import (
	"encoding/json"
	"net/http"
	"net/mail"
	"unicode"

	"gorm.io/gorm"

	"github.com/GoogleCloudPlatform/golang-samples/run/helloworld/models"
	"github.com/GoogleCloudPlatform/golang-samples/run/helloworld/util"
)

type AuthRequest struct {
	Name     string `json:"name,omitempty"`
	Email    string `json:"email"`
	Password string `json:"password"`
}
type AuthResponse struct {
	Message     string       `json:"message"`
	Token       string       `json:"token,omitempty"`
	User        *models.User `json:"user,omitempty"`
	ForgeAPIKey string       `json:"forge_api_key,omitempty"`
}

// isValidName checks if the name has between 3 and 100 characters.
func isValidName(name string) bool {
	return len(name) >= 3 && len(name) <= 100
}

// isValidEmail checks if the email follows RFC 5322 standard.
func isValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

// isValidPassword checks password complexity.
func isValidPassword(password string) (bool, string) {
	if len(password) < 8 {
		return false, "Password must be at least 8 characters long."
	}
	var (
		hasUpper, hasLower, hasNumber, hasSpecial bool
	)
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}
	if !hasUpper {
		return false, "Password must contain at least one uppercase letter."
	}
	if !hasLower {
		return false, "Password must contain at least one lowercase letter."
	}
	if !hasNumber {
		return false, "Password must contain at least one number."
	}
	if !hasSpecial {
		return false, "Password must contain at least one special character."
	}
	return true, ""
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
// @Failure 500 {string} string "Could not create user or find default plan"
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

		// Validações
		if !isValidName(req.Name) {
			http.Error(w, "Name must be between 3 and 100 characters.", http.StatusBadRequest)
			return
		}
		if !isValidEmail(req.Email) {
			http.Error(w, "Invalid email format.", http.StatusBadRequest)
			return
		}
		if valid, message := isValidPassword(req.Password); !valid {
			http.Error(w, message, http.StatusBadRequest)
			return
		}

		// Encontrar o plano "Free"
		var freePlan models.Plan
		if err := db.Where("name = ?", "Free").First(&freePlan).Error; err != nil {
			http.Error(w, "Could not find default plan", http.StatusInternalServerError)
			return
		}

		user := &models.User{
			Name:     req.Name,
			Email:    req.Email,
			Password: req.Password,
			PlanID:   freePlan.ID,
		}

		// O hook BeforeCreate irá gerar a API key e hashear a senha
		if err := db.Create(user).Error; err != nil {
			http.Error(w, "Could not create user: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Carregar o usuário com o plano para a resposta
		db.Preload("Plan").First(&user, "id = ?", user.ID)

		// Não retornar a senha e inicializar Projects como array vazio
		user.Password = ""
		if user.Projects == nil {
			user.Projects = make([]models.Project, 0)
		}

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
// @Description Authenticates a user using email and password, then returns a JWT token.
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

		// Verifica se o email foi fornecido
		if req.Email == "" {
			http.Error(w, "Email is required", http.StatusBadRequest)
			return
		}

		var user models.User
		if err := db.Preload("Plan").Preload("Projects").First(&user, "email = ?", req.Email).Error; err != nil {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}

		// Verifica a senha
		if !user.CheckPassword(req.Password) {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}

		// Gera o token JWT
		token, err := util.GenerateJWT(&user)
		if err != nil {
			http.Error(w, "Could not generate token", http.StatusInternalServerError)
			return
		}

		// Não retornar a senha e inicializar Projects como array vazio
		user.Password = ""
		if user.Projects == nil {
			user.Projects = make([]models.Project, 0)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(AuthResponse{
			Message:     "Logged in successfully",
			Token:       token,
			User:        &user,
			ForgeAPIKey: user.ForgeAPIKey,
		})
	}
}
